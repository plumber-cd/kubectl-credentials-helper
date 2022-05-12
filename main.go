package main

import (
	"log"

	"github.com/plumber-cd/kubectl-credentials-helper/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
