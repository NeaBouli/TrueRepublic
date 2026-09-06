package main

import (
	"os"

	"truerepublic/toolbootstrapevidence"
)

func main() {
	os.Exit(toolbootstrapevidence.Run(os.Args[1:], os.Stdout, os.Stderr))
}
