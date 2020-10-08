package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/manifoldco/promptui"
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

var (
	Red     = Color("\033[1;31m%s\033[0m")
	Green   = Color("\033[01;32m%s\033[0m")
	Blue    = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func (i *Items) AddTo(title, key string) []Item {
	i.Items = append(i.Items, Item{
		Type:         "default",
		Title:        title,
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
		fmt.Println(Red(err))
		return
	}
	defer db.Close()

	if len(args) == 0 {
		prompt()
		return
	}

	if args[0] == "help" {
		fmt.Printf("%s\n", Green("WELCOME TO GOTP"))
		fmt.Printf("%s\n\n", "The purpose of this package is to help with getting a 2FA otp without having to go through the shlep of taking your phone out and typing in the otp (though it's probably meant to be tedious... security via tediousnessness).")
		fmt.Printf("%s\n%s\n%s\n%s\n%s\n\n", Green("ADD A NEW TOKEN"), Magenta("gotp add key token."), "eg:", Blue("gotp add key 564HJKHJKHKKKHGJKHJKYUFHFGHJ65E"), Blue("gotp add key - local 564HJKHJKHKKKHGJKHJKYUFHFGHJ65E"))
		fmt.Printf("%s\n%s\n%s\n%s\n%s\n\n", Green("LIST KEYS"), Blue("gotp ls"), "And example of the output would be:", Magenta("1: key"), Magenta("2: key - local"))
		fmt.Printf("%s\n%s\n%s\n%s\n%s\n%s\n\n", Green("GET OTP"), Magenta("gotp key"), Magenta("gotp get key"), "or you could use the index from gotp ls", Blue("gotp 2"), "The main difference between the above commands is that get is a simple returning/printing of the key where without get the key is added to your clipboard and prints out a statement of the key being added to your clipboard. The simplified version (with the get key word) is for piping to other utilities.")
		fmt.Printf("%s\n%s\n%s %s %s\n\n", Green("GET OTP VIA PROMPT"), Magenta("gotp"), "If you only type", Blue("gotp"), "a prompt will be brought up of the keys that have been stored. You can use this prompt to select the key you want to use.")
		fmt.Printf("%s\n%s\n%s\n%s\n%s\n", Green("REMOVE KEYS"), Magenta("gotp rm key"), Magenta("gotp rm key - local"), "or you could use the index from gotp ls", Blue("gotp rm 2"))
		return
	}

	if index, err := strconv.Atoi(args[0]); err == nil {
		getOtpByIndex(index)
		return
	}

	if args[0] == "add" {
		l := len(args)
		addToken(strings.Join(args[1:(l-1)], " "), args[l-1])
		return
	}

	if args[0] == "ls" {
		listKeys()
		return
	}

	if args[0] == "rm" {
		if index, err := strconv.Atoi(args[1]); err == nil {
			removeKeyByIndex(index)
			return
		}

		l := len(args)
		removeKey(strings.Join(args[1:l], " "))
		return
	}

	if args[0] == "get" {
		l := len(args)
		getOtp(strings.Join(args[1:l], " "), "simple")
		return
	}

	if args[0] == "lsJson" {
		listJson()
		return
	}

	if args[0] == "queryJson" {
		l := len(args)
		queryJson(args[1:l])
		return
	}

	getOtp(strings.Join(args, " "), "terminal")
}

func prompt() {
	keys := getKeyList()

	if len(keys) == 0 {
		fmt.Printf("%s\n%s\n%s", Magenta("You don't have any otp's saved yet."), "Try adding one by typing:", Blue("gotp add key token"))
		return
	}

	prompt := promptui.Select{
		Label: "Select a key",
		Items: keys,
	}

	_, key, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", Red(err))
		return
	}

	getOtp(key, "terminal")
}

func getKeyList() []string {
	var keys []string
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			keys = append(keys, string(item.Key()))
		}
		return nil
	})

	if err != nil {
		fmt.Println(Red(err))
		return []string{}
	}

	return keys
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
		fmt.Println(Red(err))
	}
}

func addToken(key string, secret string) {
	txn := db.NewTransaction(true)
	defer txn.Discard()

	// Use the transaction...
	err := txn.Set([]byte(key), []byte(secret))
	if err != nil {
		fmt.Println(Red(err))
	}

	// Commit the transaction and check for error.
	if err := txn.Commit(); err != nil {
		fmt.Println(Red(err))
		return
	}

	fmt.Println("Your key and secret have been saved")
}

func removeKey(key string) {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if err != nil {
			return err
		}

		fmt.Printf("%s has been removed", Green(key))

		return nil
	})

	if err != nil {
		fmt.Println(Red(err))
	}
}

func removeKeyByIndex(index int) {
	err := db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		i := 1
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			if index != i {
				i++
				continue
			}

			item := it.Item()
			err := txn.Delete(item.Key())
			if err != nil {
				return err
			}

			fmt.Printf("%s has been removed", Green(string(item.Key())))

			break
		}

		return nil
	})

	if err != nil {
		fmt.Println(Red(err))
	}
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

		if output == "simple" {
			fmt.Println(otp)
		} else {
			fmt.Printf("Your otp is: %s\n", Green(otp))

			//Checking if the pbcopy command is on the system.
			_, err := exec.LookPath("pbcopy")
			if err != nil {
				return nil
			}

			//Copies the otp generated to your clipboard
			err = exec.Command("bash", "-c", fmt.Sprintf("echo %s | tr -d \"\n, \" | pbcopy", otp)).Run()
			if err != nil {
				return err
			}

			fmt.Printf("%s has been copied to your clipboard\n", Green(otp))
		}

		return nil
	})

	if err != nil {
		fmt.Println(Red(err))
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
		fmt.Println(Magenta("Could not find your key"))
		return
	}

	if err != nil {
		fmt.Println(Red(err))
		return
	}

	getOtp(key, "terminal")
}

func listJson() {
	items := new(Items)

	err := db.View(func(txn *badger.Txn) error {
		var k string
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k = string(item.Key())
			items.AddTo(k, k)
		}
		return nil
	})

	if err != nil {
		fmt.Println(Red(err))
		return
	}

	b, err := json.Marshal(items)
	if err != nil {
		fmt.Println(Red(err))
		return
	}
	fmt.Println(string(b))
}

func queryJson(query []string) {
	var key string
	var action string
	
	items := new(Items)
	keys := getKeyList()
	l := len(query)

	if l > 0 {
		action = query[0]
	}

	matchAdd, _ := regexp.MatchString("^ad?d?", action)
	if matchAdd {
		items.AddTo("Add", strings.Join(query, " "))
	}

	matchRm, _ := regexp.MatchString("^rm?", action)
	if matchRm {
		for _, k := range keys {
			key = string(k)
			if l == 1 || testQuery(query[1:], key) {
				items.AddTo("Remove "+key, "rm "+key)
			}
		}
	}

	if l == 0 || (!matchAdd && !matchRm) {
		for _, k := range keys {
			key = string(k)
			if l == 0 || testQuery(query, key) {
				items.AddTo(key, key)
			}
		}
	}

	b, err := json.Marshal(items)
	if err != nil {
		fmt.Println(Red(err))
		return
	}

	fmt.Println(string(b))
}

func testQuery(query []string, key string) bool {
	for _, q := range query {
		if !strings.Contains(key, q) {
			return false
		}
	}

	return true
}

func initDb() (*badger.DB, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := usr.HomeDir + "/.config/gotp"
	os.MkdirAll(dbPath, os.ModePerm)

	options := badger.DefaultOptions(dbPath)
	options.Logger = nil

	db, err := badger.Open(options)

	return db, err
}

// The code below has been taken from https://blog.gojekengineering.com/a-diy-two-factor-authenticator-in-golang-32e5641f6ec5
// Thank you Tilak Lodha for writing the article.
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
		fmt.Println(Red(err))
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
		fmt.Println(Red(err))
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
