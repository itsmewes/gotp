package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v2"
)

var db *badger.DB

type Items struct {
	Items []Item `json:"items"`
}
type Item struct {
	Type         string `json:"type"`
	Title        string `json:"title"`
	Arg          string `json:"arg"`
	Autocomplete string `json:"autocomplete"`
}

func (i *Items) AddTo(key string) []Item {
	i.Items = append(i.Items, Item{
		Type:         "default",
		Title:        key,
		Arg:          key,
		Autocomplete: key,
	})

	return i.Items
}

func main() {
	var err error
	flag.Parse()
	args := flag.Args()

	db, err = initDb()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	if len(args) == 0 {
		fmt.Println(errors.New("Please add something, anything... I don't know what you want from me."))
		return
	}

	if index, err := strconv.Atoi(args[0]); err == nil {
		getOtpByIndex(index)
		return
	}

	if args[0] == "add" {
		addToken(args[1], args[2])
		return
	}

	if args[0] == "ls" {
		listKeys()
		return
	}

	if args[0] == "lsJson" {
		listJson()
		return
	}

	if args[0] == "get" {
		getOtp(args[1], "simple")
		return
	}

	getOtp(args[0], "terminal")
}

func listKeys() {
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		i := 1
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			fmt.Printf("%d: %s\n", i, k)
			i++
		}
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
}

func listJson() {
	items := new(Items)
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			items.AddTo(string(k))
		}
		return nil
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	b, err := json.Marshal(items)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
}

func addToken(key string, secret string) {
	txn := db.NewTransaction(true)
	defer txn.Discard()

	// Use the transaction...
	err := txn.Set([]byte(key), []byte(secret))
	if err != nil {
		fmt.Println(err)
	}

	// Commit the transaction and check for error.
	if err := txn.Commit(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Your key and secret have been saved")
}

func getOtp(key string, output string) {
	err := db.View(func(txn *badger.Txn) error {
		var token string
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			token = string(val)
			return nil
		})
		if err != nil {
			return err
		}

		otp := getTOTPToken(token)

		if output == "terminal" {
			fmt.Println("Your otp is:" + otp)
			fmt.Println(otp + " has been copied to your clipboard")
		} else {
			fmt.Println(otp)
		}

		//Copies the otp generated to your clipboard
		err = exec.Command("bash", "-c", fmt.Sprintf("echo %s | tr -d \"\n, \" | pbcopy", otp)).Run()
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
}

func getOtpByIndex(index int) {
	var key string
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		i := 1
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			if index != i {
				i++
				continue
			}

			item := it.Item()
			key = string(item.Key())

			it.Close()
		}

		return nil
	})

	if key == "" {
		fmt.Println("Could not find your key")
		return
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	getOtp(key, "terminal")
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
	if err != nil {
		fmt.Println(err)
	}
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

	if err != nil {
		fmt.Println(err)
	}
	//Ignore most significant bits as per RFC 4226.
	//Takes division from one million to generate a remainder less than < 7 digits
	h12 := (int(header) & 0x7fffffff) % 1000000

	//Converts number as a string
	otp := strconv.Itoa(int(h12))

	return prefix0(otp)
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

func initDb() (*badger.DB, error) {
	options := badger.DefaultOptions("/tmp/gotp")
	options.Logger = nil

	db, err := badger.Open(options)

	return db, err
}
