// The weather CLI returns a brief weather forecast.
package main

import (
	"fmt"
	"os"
	"weather"
)

func main() {
	err := weather.RunCLI(os.Args, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
