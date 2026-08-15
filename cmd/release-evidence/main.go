package main

import (
	"fmt"
	"os"

	"truerepublic/releaseevidence"
)

func main() {
	if err := releaseevidence.NewCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
