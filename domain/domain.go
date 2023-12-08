package domain

import (
	"net/url"
	"strings"
)

func ExtractFileNameFromURL(url url.URL) string {
	slice := strings.Split(url.Path, "/")
	lastIndex := len(slice) - 1
	fName := slice[lastIndex]
	return fName
}
