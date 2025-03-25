package main

import (
	"testing"
)

// test for all of the banned words
func TestRemoveProfanity(t *testing.T) {
	testString := "This is a kerfuffle opinion I need to share with the world"
	expectedResult := "This is a **** opinion I need to share with the world"
	result := removeProfanity(testString)
	if expectedResult != result {
		t.Errorf(`removeProfanity(%q) = %q, want match for %q`, testString, result, expectedResult)
	}
}

func TestRemoveProfanityAll(t *testing.T) {
	testString := "Kerfuffle Sharbert Fornax"
	expectedResult := "**** **** ****"
	result := removeProfanity(testString)
	if expectedResult != result {
		t.Errorf(`removeProfanity(%q) = %q, want match for %q`, testString, result, expectedResult)
	}
}

func TestRemoveProfanityPunctuation(t *testing.T) {
	testString := "Kerfuffle! Sharbert Fornax"
	expectedResult := "Kerfuffle! **** ****"
	result := removeProfanity(testString)
	if expectedResult != result {
		t.Errorf(`removeProfanity(%q) = %q, want match for %q`, testString, result, expectedResult)
	}
}
