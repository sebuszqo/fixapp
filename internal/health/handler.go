package health

import (
	"fixapp/pkg/ctxlog"
	"fixapp/pkg/response"
	"net/http"
	"time"
)

type Handler struct {
	started time.Time
	version string
}

// HealthResponse represents the health check response
type healthResponse struct {
	Status  string    `json:"status" example:"ok"`
	Version string    `json:"version" example:"1.0.0"`
	Started time.Time `json:"started" example:"2025-10-26T10:30:00Z"`
	Uptime  string    `json:"uptime" example:"5h30m"`
}

// ReadinessResponse represents the readiness check response
type readinessResponse struct {
	Status string `json:"status" example:"ready"`
}

func New(version string) *Handler {
	return &Handler{
		started: time.Now(),
		version: version,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.healthCheck)
	mux.HandleFunc("GET /ready", h.readinessCheck)
}

// healthCheck godoc
// @Summary      Health check
// @Description  Check if the application is running
// @Tags         monitoring
// @Accept       json
// @Produce      json
// @Success      200  {object}  healthResponse
// @Router       /healthz [get]
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())
	log.Info("health check")

	response.JSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Version: h.version,
		Started: h.started,
		Uptime:  time.Since(h.started).String(),
	})
}

// readinessCheck godoc
// @Summary      Readiness check
// @Description  Check if the application is ready to accept traffic
// @Tags         monitoring
// @Accept       json
// @Produce      json
// @Success      200  {object}  readinessResponse
// @Router       /ready [get]
func (h *Handler) readinessCheck(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())
	log.Info("readiness check")

	response.JSON(w, http.StatusOK, readinessResponse{
		Status: "ready",
	})
}
