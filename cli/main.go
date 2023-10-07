package main

import (
	"bz/commands"
	"os"

	log "k8s.io/klog/v2"
)

func main() {
	_, exists := os.LookupEnv("BAAZ_URL")
	if !exists {
		log.Error("Env BAAZ_URL does not exist")
		os.Exit(1)
	}
	// _, exists := os.LookupEnv("BAAZ_USERNAME")
	// if !exists {
	// 	os.Exit(1)
	// }
	// _, exists := os.LookupEnv("BAAZ_PASSWORD")
	// if !exists {
	// 	os.Exit(1)
	// }

	err := commands.Execute()
	if err != nil && err.Error() != "" {
		log.Error(err)
		os.Exit(1)
	}
}
