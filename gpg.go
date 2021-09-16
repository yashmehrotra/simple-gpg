package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

# To encrypt a file
$ simple-gpg accounts.pdf
Enter password:

# To encrypt a file with passowrd as argument
$ simple-gpg -password secret-word accounts.pdf

# To decrypt a file
$ simple-gpg -decrypt accounts.pdf

# To encrypt a folder
$ simple-gpg /path/to/folder

Note: When encrypting a folder, simple-gpg uses tar.gz format to compress the folder into an archive and then encrypts the archive
`

func EncrpytFile(file string, password []byte, algo string, outputFile string) error {
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

	w, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer w.Close()

	fmt.Println("Encrypting:", file)
	pt, err := openpgp.SymmetricallyEncrypt(w, password, nil, config)
	if _, err := io.Copy(pt, f); err != nil {
		return err
	}
	fmt.Println("Encryption successful:", outputFile)
	return pt.Close()
}

func tarDir(srcPath, dstPath string) error {

	f, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(fmt.Sprintf("Error creating file %v", err.Error()))
	}
	return Tar(srcPath, f)
}

func compressIfDir(path, suffix string) (bool, string) {
	// If given path is a directory, compress it and return
	// the path  of the compressed file
	compressed := false

	// No action required, it is not a directory
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return compressed, ""
	}

	// Tar + gzip this
	fmt.Printf("WARNING: %s is a directory. Converting into .tar.gz\n", path)
	outputPath := filepath.Join("/tmp/", path, suffix)
	err := tarDir(path, outputPath)
	if err != nil {
		// TODO: Better handling
		panic(err)
	}
	compressed = true
	return compressed, outputPath
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
	cliPassword := flag.String("password", "", "Password to use when encrypting/decrypting")
	outputFile := flag.String("outputFile", "", "Path for output file. (Default: <filename>.gpg)")

	flag.Usage = func() {
		fmt.Println(Description)
		flag.PrintDefaults()
		fmt.Println(Examples)
	}
	flag.Parse()

	var err error
	if len(flag.Args()) != 1 {
		fmt.Println("Error: Unexpected number of arguments")
		flag.Usage()
		os.Exit(1)
	}
	file := flag.Args()[0]

	var password []byte
	if *cliPassword == "" {
		fmt.Printf("Enter password: ")
		password, err = terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		fmt.Println()
	} else {
		password = []byte(*cliPassword)
	}

	if *decrypt {
		err = DecrpytFile(file, password)
	} else {
		outputPath := *outputFile
		suffixDir := "-compressed.tar.gz"
		if compressed, newPath := compressIfDir(file, suffixDir); compressed {
			if outputPath == "" {
				outputPath = file + suffixDir + ".gpg"
			}

			file = newPath
			// Remove tar, we only need to preserve the original directory
			// os.Remove returns an error but we won't check it cause its not important
			err = EncrpytFile(file, password, *algo, outputPath)
			os.Remove(file)
		} else {
			if outputPath == "" {
				outputPath = file + ".gpg"
			}
			err = EncrpytFile(file, password, *algo, outputPath)
		}
	}

	if err != nil {
		panic(err)
	}

}
