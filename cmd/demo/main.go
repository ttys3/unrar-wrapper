package main

/*
A sample program that extracts all files from a given .7z file
to current directory.
*/

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ttys3/unrar-wrapper"
)

func usageAndExit() {
	fmt.Printf("usage: test file.rar\n")
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		usageAndExit()
	}
	path := os.Args[1]
	a, err := unrarwrapper.NewArchive(path)
	if err != nil {
		fmt.Printf("lzmadec.NewArchive('%s') failed with '%s'\n", path, err)
		os.Exit(1)
	}
	fmt.Printf("opened archive '%s'\n", path)
	fmt.Printf("Extracting %d entries\n", len(a.Entries))
	for _, e := range a.Entries {
		dirname := filepath.Dir(e.Name)
		if _, err := os.Stat(dirname); os.IsNotExist(err) {
			os.MkdirAll(dirname, 0755)
		}
		err = a.ExtractToFile(e.Name, e.Name)
		if err != nil {
			fmt.Printf("a.ExtractToFile('%s') failed with '%s'\n", e.Name, err)
			os.Exit(1)
		}
		fmt.Printf("Extracted '%s'\n", e.Name)
	}
}
