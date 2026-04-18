package handler

import (
	"net/http"

	"github.com/Bharat1Rajput/My-Portfolio/web/templates/pages"
)

// ContactHandler serves GET /contact
type ContactHandler struct{}

func NewContactHandler() *ContactHandler {
	return &ContactHandler{}
}

func (h *ContactHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pages.Contact().Render(r.Context(), w)
}
