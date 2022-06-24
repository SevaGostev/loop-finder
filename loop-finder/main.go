package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)


func main () {

	if len(os.Args) < 2 {
		log.Fatal("Specify input files")
	}

	if os.Args[1] == "-i" {

		if len(os.Args) == 2 {
			interactive("")
		} else {
			if _, err := os.Stat(os.Args[2]); err != nil {
				interactive("")
			} else {
				interactive(os.Args[2])
			}
		}
	} else {
		nonInteractive()
	}
}

func interactive(baseDir string) {

	fmt.Print("Interactive mode\n")

	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		input, err := r.ReadString('\n')

		if err != nil {
			break
		}

		input = strings.TrimSuffix(input, "\n")

		if input == "" || input == "q" {
			break
		}

		tokens := strings.Split(input, ",")

		if len(tokens) < 2 {
			handleFile(filepath.Join(baseDir, input), 0)
		} else {
			hint, err := strconv.ParseUint(tokens[1], 10, 64)

			if err != nil {
				fmt.Print("Could not parse hint.\n")
				continue
			}

			handleFile(filepath.Join(baseDir, tokens[0]), hint)
		}
	}

	os.Exit(0)
}

func nonInteractive() {

	for i := 1; i < len(os.Args); i++ {

		handleFile(os.Args[i], 0)
	}

	os.Exit(0)
}