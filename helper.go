package main

import (
	"golang.org/x/net/html"
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

func NextSiblings(node *html.Node, amount int) *html.Node {
	sibling := node.NextSibling

	for i := 1; i < amount; i++ {
		sibling = sibling.NextSibling
	}

	return sibling
}

func PrevSiblings(node *html.Node, amount int) *html.Node {
	sibling := node.PrevSibling

	for i := 1; i < amount; i++ {
		sibling = sibling.PrevSibling
	}

	return sibling
}
