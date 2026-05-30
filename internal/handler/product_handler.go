package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/raymond/wyzauto-project/internal/service"
)

type ProductHandler struct {
	builder *service.ProductDocumentBuilder
}

func NewProductHandler(builder *service.ProductDocumentBuilder) *ProductHandler {
	return &ProductHandler{builder: builder}
}

func (h *ProductHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /products/{id}/document", h.buildDocument)
}

func (h *ProductHandler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProductHandler) buildDocument(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimSpace(r.PathValue("id"))
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "product id is required"})
		return
	}

	document, err := h.builder.Build(r.Context(), productID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, document)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
