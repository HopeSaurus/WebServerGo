package main

import "strings"

func removeProfanity(s string) string {
	bannedWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	msgArr := strings.Split(s, " ")
	for i, word := range msgArr {
		if bannedWords[strings.ToLower(word)] {
			msgArr[i] = "****"
		}
	}
	return strings.Join(msgArr, " ")
}
