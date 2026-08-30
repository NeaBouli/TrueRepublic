package main

import (
	"os"

	"truerepublic/candidateevidence"
)

func main() {
	os.Exit(candidateevidence.Run(os.Args[1:], os.Stdout, os.Stderr))
}
