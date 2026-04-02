package main

import (
	"flag"
	"fmt"
	"strings"
)

func main() {
	omitNewline := flag.Bool("n", false, "Omit the trailing newline")
	flag.Parse()

	args := flag.Args()
	output := strings.Join(args, " ")

	if *omitNewline {
		fmt.Print(output)
	} else {
		fmt.Println(output)
	}
}
