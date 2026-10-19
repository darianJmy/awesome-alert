package main

import (
	"aiops-backend/config"
	"aiops-backend/database"
	"aiops-backend/handlers"
	"aiops-backend/routes"
	"log"
)

func main() {
	database.Init()
	handlers.StartScheduler()

	r := routes.SetupRouter()

	log.Println("server starting on", config.AppConfig.ServerPort)
	if err := r.Run(config.AppConfig.ServerPort); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
