package main

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/examples/trace-app/internal/controllers"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

func main() {
	currDir, err := utils.GetCurrentDirectory()
	if err != nil {
		logger.Panic(err)
	}
	configPath := currDir + "/examples/trace-app/configs/app.yaml"
	config := raiden.LoadConfig(&configPath)

	controllerRegistry := raiden.NewControllerRegistry()
	controllerRegistry.Register(
		controllers.HelloWordController,
		controllers.GetPostController,
	)

	server := raiden.NewServer(config, controllerRegistry.Controllers)
	server.Run()
}
