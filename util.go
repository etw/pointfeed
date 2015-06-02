package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
)

func getRid() (*string, error) {
	ridbin := make([]byte, 8)
	_, err := rand.Read(ridbin)
	if err != nil {
		log.Println("[Error] Couldn't generate request id: %s", err)
		return nil, errors.New("Couldn't generate request id")
	}
	ridstr := base64.URLEncoding.EncodeToString(ridbin)
	return &ridstr, nil
}

func findNl(str []rune) int {
	for i, c := range str {
		if c == '\n' {
			return i
		}
	}
	return -1
}
