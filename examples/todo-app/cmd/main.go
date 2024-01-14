package main

import (
	"fmt"
	"os"

	"github.com/sev-2/raiden/pkg/raiden"
)

func main() {
	// load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		raiden.Panic(err)
	}
	configFolder := fmt.Sprintf("%s/examples/todo-app/configs/app.yaml", currentDir)
	config := raiden.LoadConfig(&configFolder)

	// Setup server
	app := raiden.NewServer(config)
	app.Run()
}
