package api

import (
	"net/http"

	"github.com/jrapoport/gothic/conf"
)

func (a *API) handleHealthCheck(w http.ResponseWriter, r *http.Request) error {
	return sendJSON(w, http.StatusOK, a.HealthCheck())
}

func (a *API) HealthCheck() map[string]string {
	return map[string]string{
		"version":     conf.CurrentVersion(),
		"name":        "Gothic",
		"description": "Gothic is a user registration and authentication API",
	}
}
