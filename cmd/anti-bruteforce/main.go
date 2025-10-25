package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Alexandr-Snisarenko/Otus-Anti-bruteforce/internal/config"
	"github.com/Alexandr-Snisarenko/Otus-Anti-bruteforce/internal/logger"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/anti-bruteforce/config.yaml", "Path to configuration file")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "anti-bruteforce exited with error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return nil
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}

	logg := logger.New(&cfg.Logger)
	log.Printf("Logger initialized with level: %s", cfg.Logger.Level)

	logg.Info("Application started")

	// Application logic would go here...

	logg.Info("Application stopped")

	return nil
}
