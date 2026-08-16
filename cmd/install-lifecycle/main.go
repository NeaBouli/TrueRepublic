package main

import (
	"fmt"
	"os"

	"truerepublic/installlifecycle"
)

func main() {
	if err := installlifecycle.NewCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
