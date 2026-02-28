package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("error") // OK - in main
	os.Exit(1)       // OK - in main
}