# gotp
Handling Google Authenitcator codes in your terminal

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
