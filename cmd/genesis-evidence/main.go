package main

import (
	"fmt"
	"os"

	"truerepublic/genesisevidence"
)

func main() {
	if err := genesisevidence.NewCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
