package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v2"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		check(errors.New("Please add something, anything... I don't know what you want from me."))
		return
	}

	if args[0] == "add" {
		addToken(args[1], args[2])
		return
	}

	getOtp(args[0])
}

func addToken(key string, secret string) {
	db, err := badger.Open(badger.DefaultOptions("/tmp/gotp"))
	if err != nil {
		check(err)
	}
	defer db.Close()

	txn := db.NewTransaction(true)
	defer txn.Discard()

	// Use the transaction...
	err = txn.Set([]byte(key), []byte(secret))
	if err != nil {
		check(err)
	}

	// Commit the transaction and check for error.
	if err := txn.Commit(); err != nil {
		check(err)
		return
	}

	fmt.Println("Your key and secret have been saved")
}

func getOtp(key string) {
	db, err := badger.Open(badger.DefaultOptions("/tmp/gotp"))
	if err != nil {
		check(err)
	}
	defer db.Close()

	err = db.View(func(txn *badger.Txn) error {
		token, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		otp := getTOTPToken(token.String())

		fmt.Println("Your otp is:" + otp)
		fmt.Println(otp + " has been copied to your clipboard")

		//Copies the otp generated to your clipboard
		err = exec.Command("bash", "-c", fmt.Sprintf("echo %s | tr -d \"\n, \" | pbcopy", otp)).Run()
		check(err)

		return nil
	})

	check(err)
}

func getTOTPToken(secret string) string {
	//The TOTP token is just a HOTP token seeded with every 30 seconds.
	interval := time.Now().Unix() / 30
	return getHOTPToken(secret, interval)
}

func getHOTPToken(secret string, interval int64) string {

	//Converts secret to base32 Encoding. Base32 encoding desires a 32-character
	//subset of the twenty-six letters A–Z and ten digits 0–9
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	check(err)
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(interval))

	//Signing the value using HMAC-SHA1 Algorithm
	hash := hmac.New(sha1.New, key)
	hash.Write(bs)
	h := hash.Sum(nil)

	// We're going to use a subset of the generated hash.
	// Using the last nibble (half-byte) to choose the index to start from.
	// This number is always appropriate as it's maximum decimal 15, the hash will
	// have the maximum index 19 (20 bytes of SHA1) and we need 4 bytes.
	o := (h[19] & 15)

	var header uint32
	//Get 32 bit chunk from hash starting at the o
	r := bytes.NewReader(h[o : o+4])
	err = binary.Read(r, binary.BigEndian, &header)

	check(err)
	//Ignore most significant bits as per RFC 4226.
	//Takes division from one million to generate a remainder less than < 7 digits
	h12 := (int(header) & 0x7fffffff) % 1000000

	//Converts number as a string
	otp := strconv.Itoa(int(h12))

	return prefix0(otp)
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

//Append extra 0s if the length of otp is less than 6
//If otp is "1234", it will return it as "001234"
func prefix0(otp string) string {
	if len(otp) == 6 {
		return otp
	}
	for i := (6 - len(otp)); i > 0; i-- {
		otp = "0" + otp
	}
	return otp
}
