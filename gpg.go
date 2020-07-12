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
		fmt.Printf("A tool to simplify gpg\n\n")
		fmt.Printf("simple-gpg [args] file\n\n")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Printf("\n\nExamples:\n\n")
		fmt.Println("To encrypt a file")
		fmt.Printf("$ simple-gpg accounts.pdf\n\n")
		fmt.Println("To decrypt a file")
		fmt.Printf("$ simple-gpg -decrypt accounts.pdf\n")
	}
	flag.Parse()

	if len(flag.Args()) != 1 {
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
		err = EncrpytFile(file, password, *algo)
	}

	if err != nil {
		panic(err)
	}

}
