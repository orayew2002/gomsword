package main

import (
	"context"
	"fmt"
	"os"

	"github.com/orayew2002/gomsword/pkg/wordtmpl"
)

// main is the CLI entry point for printing extracted keys.
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: wordkeys <file.docx|file.doc>")
		os.Exit(2)
	}

	keys, err := wordtmpl.ExtractKeys(context.Background(), os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	for _, k := range keys {
		fmt.Println(k)
	}
}
