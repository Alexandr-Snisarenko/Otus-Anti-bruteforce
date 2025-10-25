package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	if flag.Arg(0) == "version" {
		printVersion()
	}

	logger := log.New(os.Stdout, "anti-bruteforce: ", log.LstdFlags)
	logger.Println("Anti-Bruteforce service started")

	// Здесь будет основная логика сервиса
}
