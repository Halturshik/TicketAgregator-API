package health

import (
	"net/http"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

const (
	statusAlive    = "alive"
	statusReady    = "ready"
	statusNotReady = "not_ready"
)

type Handler struct {
	checker *Checker
}

type response struct {
	Status string           `json:"status"`
	Checks map[string]State `json:"checks,omitempty"`
}

func New(timeout time.Duration, dependencies ...Dependency) *Handler {
	return &Handler{checker: NewChecker(timeout, dependencies...)}
}

func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	_ = httpx.WriteJSON(w, http.StatusOK, response{Status: statusAlive})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	report := h.checker.Check(r.Context())
	checks := make(map[string]State, len(report.Checks))
	for name, result := range report.Checks {
		checks[name] = result.State
	}

	status := http.StatusOK
	resultStatus := statusReady
	if !report.Ready {
		status = http.StatusServiceUnavailable
		resultStatus = statusNotReady
	}
	w.Header().Set("Cache-Control", "no-store")
	_ = httpx.WriteJSON(w, status, response{Status: resultStatus, Checks: checks})
}
