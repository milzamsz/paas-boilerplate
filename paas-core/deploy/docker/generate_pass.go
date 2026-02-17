package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	pass := "admin"
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Admin Hash: %s\n", string(hash))

	passDemo := "demo"
	hashDemo, err := bcrypt.GenerateFromPassword([]byte(passDemo), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Demo Hash: %s\n", string(hashDemo))
}
