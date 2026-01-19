package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	Version   = "dev"
	BuildTime = ""
)

func main() {
	// Handle version flag
	versionFlag := flag.Bool("version", false, "Print version info and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Printf("Version:   %s\n", Version)
		fmt.Printf("BuildTime: %s\n", BuildTime)
		os.Exit(0)
	}

	// Create and run the application
	log.Println("🏗️  Initializing UGCL Multi-Service Platform...")

	app := NewApplication()
	app.Run()
}
