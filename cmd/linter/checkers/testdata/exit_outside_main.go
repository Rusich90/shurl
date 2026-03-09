package main

import (
	"log"
	"os"
)

func foo() {
	log.Fatal("error") // want "вызов log.Fatal вне функции main"
}

func bar() {
	os.Exit(1) // want "вызов os.Exit вне функции main"
}
