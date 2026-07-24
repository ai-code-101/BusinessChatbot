package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("rr22147rr"), bcrypt.DefaultCost)
	fmt.Println(string(hash))
}