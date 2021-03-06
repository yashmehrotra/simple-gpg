package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/crypto/ssh/terminal"
)

const Description = `
A tool to simplify gpg

simple-gpg [args] file

Flags:
`

const Examples = `
Examples:

To encrypt a file
$ simple-gpg accounts.pdf

To decrypt a file
$ simple-gpg -decrypt accounts.pdf

To encrypt a folder
$ simple-gpg /path/to/folder

Note: When encrypting a folder, simple-gpg uses tar.gz format to compress the folder into an archive and then encrypts the archive
`

func EncrpytFile(file string, password []byte, algo string) error {
	cipher := packet.CipherAES256
	if algo == "AES" {
		cipher = packet.CipherAES128
	} else if algo == "AES192" {
		cipher = packet.CipherAES192
	} else if algo != "AES256" {
		fmt.Println("Unknown cipher", algo, "provided. Using AES256 as default")
	}

	config := &packet.Config{
		DefaultCipher: cipher,
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := os.Create(file + ".gpg")
	if err != nil {
		return err
	}
	defer w.Close()

	fmt.Println("Encrypting:", file)
	pt, err := openpgp.SymmetricallyEncrypt(w, password, nil, config)
	if _, err := io.Copy(pt, f); err != nil {
		return err
	}
	fmt.Println("Encryption successful:", file+".gpg")
	return pt.Close()
}

func tarDir(srcPath, dstPath string) error {

	f, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(fmt.Sprintf("Error creating file %v", err.Error()))
	}
	return Tar(srcPath, f)
}

func compressIfDir(path string) (bool, string) {
	// If given path is a directory, compress it and return
	// the path  of the compressed file
	compressed := false

	// No action required, it is not a directory
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return compressed, ""
	}

	// Tar + gzip this
	fmt.Printf("WARNING: %s is a directory. Converting into .tar.gz\n", path)
	suffix := "-compressed.tar.gz"
	err := tarDir(path, path+suffix)
	if err != nil {
		// TODO: Better handling
		panic(err)
	}
	compressed = true
	return compressed, path + suffix
}

func DecrpytFile(file string, password []byte) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	failed := false
	prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if failed {
			return nil, errors.New("Decryption failed. Check password")
		}
		failed = true
		return password, nil
	}

	fmt.Println("Decrypting:", file)
	md, err := openpgp.ReadMessage(f, nil, prompt, &packet.Config{})
	if err != nil {
		return err
	}
	plaintext, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("decrypted-%s", strings.TrimSuffix(file, ".gpg"))
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	dst.Write(plaintext)
	fmt.Println("Decryption successful:", path)
	return nil

}

func main() {
	algo := flag.String("cipher-algo", "AES256", "Cipher algorithm to be used. Choose one of AES, AES192, AES256")
	decrypt := flag.Bool("decrypt", false, "Set to true if you want to decrypt a file")

	flag.Usage = func() {
		fmt.Println(Description)
		flag.PrintDefaults()
		fmt.Println(Examples)
	}
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("Error: Unexpected number of arguments")
		flag.Usage()
		os.Exit(1)
	}
	file := flag.Args()[0]

	fmt.Printf("Enter password: ")
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	fmt.Println()

	if *decrypt {
		err = DecrpytFile(file, password)
	} else {
		if compressed, newPath := compressIfDir(file); compressed {
			file = newPath
			// Remove tar, we only need to preserve the original directory
			// os.Remove returns an error but we won't check it cause its not important
			err = EncrpytFile(file, password, *algo)
			os.Remove(file)
		} else {
			err = EncrpytFile(file, password, *algo)
		}
	}

	if err != nil {
		panic(err)
	}

}
