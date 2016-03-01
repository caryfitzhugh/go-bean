package main

import (
	"os"
)

func check_err(err error, message string) {
	if err != nil {
		println("ERROR")
		println(message)
		println("")
		os.Exit(1)
	}
}
