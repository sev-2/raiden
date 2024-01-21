package main

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

func main() {
	currDir, err := utils.GetCurrentDirectory()
	if err != nil {
		logger.Panic(err)
	}
	configPath := currDir + "/examples/todo-app/configs/app.yaml"
	config := raiden.LoadConfig(&configPath)

	// Setup server
	app := raiden.NewServer(config)
	app.Run()
}
