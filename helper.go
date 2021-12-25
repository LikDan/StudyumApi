package main

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"strings"
)

func normalizeStr(str string) string {
	return strings.Title(strings.Trim(str, " "))
}

func checkError(err error) bool {
	if err != nil {
		log.Fatal(err)
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

func getUrlData(r *http.Request, key string) (string, error) {
	keys, ok := r.URL.Query()[key]

	if !ok || len(keys[0]) < 1 {
		return "", fmt.Errorf("url Param %s is missing", key)
	}

	return keys[0], nil
}
