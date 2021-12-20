package main

import (
	"log"
	"strings"
)

func normalizeStr(str string) string {
	return strings.Title(strings.Trim(str, " "))
}

func checkError(err error, exit bool) bool {
	if err != nil {
		if exit {
			log.Print(err)
			return true
		} else {
			log.Fatal(err)
		}
	}
	return false
}
