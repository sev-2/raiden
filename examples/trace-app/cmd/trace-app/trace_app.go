package main

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/examples/trace-app/internal/routes"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

func main() {
	// load configuration
	currDir, err := utils.GetCurrentDirectory()
	if err != nil {
		logger.Panic(err)
	}
	configPath := currDir + "/examples/trace-app/configs/app.yaml"
	config := raiden.LoadConfig(&configPath)

	// create new server
	server := raiden.NewServer(config)

	// register route
	routes.RegisterRoute(server)

	// run server
	server.Run()
}
