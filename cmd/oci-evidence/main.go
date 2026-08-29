package main

import (
	"os"

	"truerepublic/ocievidence"
)

func main() {
	os.Exit(ocievidence.Run(os.Args[1:], os.Stdout, os.Stderr))
}
