package main

import (
	"net/http"
)

// swagger:model healthCheckResponse
type HealthCheckResponse struct {
	Status      string `json:"status" example:"available"`
	Environment string `json:"environment" example:"development"`
	Version     string `json:"version" example:"1.0.0"`
}

// swagger:route GET /healthcheck healthcheck healthcheckEndpoint
// Health check endpoint.
// Checks if the application is running.
// responses:
//
//	200: healthCheckResponse
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
