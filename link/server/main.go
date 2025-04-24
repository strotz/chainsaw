package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/strotz/chainsaw/link"
)

// Run standalone server
func main() {
	flag.Parse()
	s := link.Server{}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, os.Kill)

	if err := s.Start(); err != nil {
		log.Fatalln("Failed to start server:", err)
	}
	log.Println("Server started...")
	<-done // Wait for ctrl+c
}
