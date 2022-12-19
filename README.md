# simple-gpg
A simple GPG tool for encrypting and decrypting files

### Installation
```
go install github.com/yashmehrotra/simple-gpg@latest
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
  -password string
    	Password to use when encrypting/decrypting



Examples:

# To encrypt a file
$ simple-gpg accounts.pdf
Enter password:

# To encrypt a file with passowrd as argument
$ simple-gpg -password secret-word accounts.pdf

# To decrypt a file
$ simple-gpg -decrypt accounts.pdf

# To encrypt a folder
$ simple-gpg /path/to/folder

Note: When encrypting a folder, simple-gpg uses tar.gz format to compress the folder into an archive
and then encrypts the archive
```
