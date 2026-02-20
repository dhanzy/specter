package main

import (
	"Specter/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("specter: %v", err)
	}
}
