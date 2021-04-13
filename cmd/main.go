// The weather CLI returns a brief weather forecast.
package main

import (
	"fmt"
	"os"
	"weather"
)

func main() {
	err := weather.RunCLI(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
