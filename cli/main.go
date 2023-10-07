package main

import (
	"bz/commands"
	"os"

	log "k8s.io/klog/v2"
)

func main() {
	err := commands.Execute()
	if err != nil && err.Error() != "" {
		log.Error(err)
		os.Exit(1)
	}
}
