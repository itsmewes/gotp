![gotp logo](https://raw.githubusercontent.com/itsmewes/gotp/master/images/logo.png "gotp logo")

# gotp
Handling Google Authenticator codes in your terminal

![Walkthrough](https://raw.githubusercontent.com/itsmewes/gotp/master/images/gotp-walkthrough.gif "Walk through")

## Getting started

### Grab the binary
The quickest way to get set up is to download the latest binary from [releases](https://github.com/itsmewes/gotp/releases)

### Build from source
This project make use of these great projects [Badger](https://github.com/dgraph-io/badger/) and [promptui](https://github.com/manifoldco/promptui).
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

Gotp tries to find the right key based on a very basic fuzzy search. For example if you have two keys, one for local and one for production you may have two entries like the following `website - local` and `website - production`, if you type `gotp web prod` gotp will get the otp for `website - production`.

### Get OTP via prompt
```
gotp 
```
If you only type `gotp` a prompt will be brought up of the keys that have been stored. You can use this prompt to select the key you want to use.

### Remove keys
```
gotp rm key
gotp rm key - local
```
or you could use the index from `gotp ls`
```
gotp rm 2
```
Gotp uses the same basic fuzzy searching for removing keys as it does for getting them.

## Usage for Alfred
There are some commands that I have added in so I can use this tool in conjunction with Alfred (a Mac app utility.) You can find the Alfred workflow on [Packal](https://www.packal.org/workflow/gotp)

### List keys, return in JSON format
```
gotp lsJson
```

### Search for keys, return in JSON format
```
gotp queryJson query
```
