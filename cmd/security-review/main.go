package main

import (
	"os"

	"truerepublic/securityreview"
)

func main() {
	os.Exit(securityreview.Run(os.Args[1:], os.Stdout, os.Stderr))
}
