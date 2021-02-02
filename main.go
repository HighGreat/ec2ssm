package main

import (
	"ec2ssm/ui"
	"log"
	"os"
)

func main() {
	if err := ui.NewUi().Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
