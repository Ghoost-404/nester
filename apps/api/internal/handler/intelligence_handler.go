package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/suncrestlabs/nester/apps/api/internal/domain/intelligence"
	"github.com/suncrestlabs/nester/apps/api/internal/service"
)

type IntelligenceHandler struct {
	prometheus *service.PrometheusClient
}

func NewIntelligenceHandler(prometheus *service.PrometheusClient) *IntelligenceHandler {
	return &IntelligenceHandler{
		prometheus: prometheus,
	}
}

func (h *IntelligenceHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/vaults/{id}/recommendations", h.GetVaultRecommendations)
	mux.HandleFunc("GET /api/v1/intelligence/market", h.GetMarketSentiment)
	mux.HandleFunc("GET /api/v1/intelligence/portfolio/{userId}", h.GetPortfolioInsights)
	mux.HandleFunc("POST /api/v1/intelligence/savings-plan", h.CreateSavingsPlan)
}

func (h *IntelligenceHandler) GetVaultRecommendations(w http.ResponseWriter, r *http.Request) {
	vaultID := chi.URLParam(r, "id")
	if vaultID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recs, err := h.prometheus.GetVaultRecommendations(r.Context(), vaultID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(recs)
}

func (h *IntelligenceHandler) GetMarketSentiment(w http.ResponseWriter, r *http.Request) {
	report, err := h.prometheus.GetMarketSentiment(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

func (h *IntelligenceHandler) GetPortfolioInsights(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	insights, err := h.prometheus.GetPortfolioInsights(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(insights)
}

func (h *IntelligenceHandler) CreateSavingsPlan(w http.ResponseWriter, r *http.Request) {
	var req intelligence.SavingsPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	plan, err := h.prometheus.CreateSavingsPlan(r.Context(), req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(plan)
}
