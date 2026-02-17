package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := map[string]string{
		"admin": "admin",
		"local": "local",
	}
	for label, pw := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Error hashing %s: %v\n", label, err)
			continue
		}
		fmt.Printf("%s: %s\n", label, string(hash))
	}
}
