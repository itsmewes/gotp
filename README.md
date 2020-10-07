![gotp logo](https://raw.githubusercontent.com/itsmewes/gotp/master/images/logo.png "gotp logo")

# gotp
Handling Google Authenitcator codes in your terminal

![Walkthrough](https://raw.githubusercontent.com/itsmewes/gotp/master/images/gotp-walkthrough.gif "Walkthrough")

## Getting started

### Grab the binary
The quickest way to get set up is to download the latest binary from [https://github.com/itsmewes/gotp/releases](releases)

### Build from source
This project make use of these great projects [https://github.com/dgraph-io/badger/](Badger) and [https://github.com/manifoldco/promptui](promptui).
Make sure to get them installed:

#### Badger
`go get github.com/dgraph-io/badger/v2`

#### promptui
`go get github.com/manifoldco/promptui`

Once you have the dependencies installed, clone the master branch to you computer:
`git clone https://github.com/itsmewes/gotp.git`

When the cloning is complete `cd` into the newly created gotp folder and run `go install .` or `go build .` to create a new binary. Make sure that your go bin folder is in your PATH or that you move the binary after building to somewhere like `/usr/local/bin` for you to easily reference it in your terminal.

## Usage

### Add a new token
```
gotp add key 564HJKHJKHKKKHGJKHJKYUFHFGHJ65E
gotp add key - local 564HJKHJKHKKKHGJKHJKYUFHFGHJ65E
```

### List keys
```
gotp ls
```
And example of the output would be:
```
1: key
2: key - local
```

### Get OTP
```
gotp key
gotp get key
```
or you could use the index from `gotp ls`
```
gotp 2
```
The main difference between the above commands is that `get` is a simple returning/printing of the key where without `get` the key is added to your clipboard and prints out a statement of the key being added to your clipboard. The simplified version (with the `get` key word) is for piping to other utilities.

### Get OTP via prompt
```
gotp 
```
If you only type `gotp` a prompt will be brought up of the keys that have been stored. You can use this prompt to select the key you want to use.
The prompt being used is "github.com/manifoldco/promptui".

### Remove keys
```
gotp rm key
gotp rm key - local
```
or you could use the index from `gotp ls`
```
gotp rm 2
```

## Usage for Alfred
There are some commands that I have added in so I can use this tool in conjuction with Alfred (a Mac app utility.)

### List keys, return in json format
```
gotp lsJson
```

### Search for keys, return in json format
```
gotp queryJson query
```
