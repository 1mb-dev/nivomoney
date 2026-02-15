package main

import (
	"net/http"

	"github.com/1mb-dev/nivomoney/services/risk/internal/handler"
	"github.com/1mb-dev/nivomoney/services/risk/internal/repository"
	"github.com/1mb-dev/nivomoney/services/risk/internal/service"
	"github.com/1mb-dev/nivomoney/shared/server"
)

func main() {
	server.Run(server.ServiceConfig{
		Name: "risk",
		SetupHandler: func(ctx *server.BootstrapContext) (http.Handler, error) {
			// Initialize repositories
			ruleRepo := repository.NewRiskRuleRepository(ctx.DB.DB)
			eventRepo := repository.NewRiskEventRepository(ctx.DB.DB)

			// Initialize services
			riskService := service.NewRiskService(ruleRepo, eventRepo)

			// Initialize router
			router := handler.NewRouter(riskService)

			return router.SetupRoutes(), nil
		},
	})
}
