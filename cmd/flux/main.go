package main

import (
	"fmt"
	"os"

	"fluxgo/internal/cli"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	var err error
	switch os.Args[1] {
	case "dev":
		err = cli.Dev()
	case "version", "--version", "-v":
		fmt.Printf("FluxGo CLI %s\n", version)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "flux: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`FluxGo CLI

Usage:
  flux <command>

Commands:
  dev       Run the application with hot reload
  version   Print the CLI version
  help      Show this help`)
}
