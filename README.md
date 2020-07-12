# gpg-util
A simple GPG tool for encrypting and decrypting files

### Installation
```
go get github.com/yashmehrotra/simple-gpg
```

### Usage
```
A tool to simplify gpg

simple-gpg [args] file

Flags:
  -cipher-algo string
    	Cipher algorithm to be used. Choose one of AES, AES192, AES256 (default "AES256")
  -decrypt
    	Set to true if you want to decrypt a file


Examples:

To encrypt a file
$ simple-gpg accounts.pdf

To decrypt a file
$ simple-gpg -decrypt accounts.pdf
```
