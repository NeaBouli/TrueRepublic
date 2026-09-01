package main

import (
	"os"

	"truerepublic/crossrunevidence"
)

func main() {
	os.Exit(crossrunevidence.Run(os.Args[1:], os.Stdout, os.Stderr))
}
