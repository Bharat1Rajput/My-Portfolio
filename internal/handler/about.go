package handler

import (
	"net/http"

	"github.com/Bharat1Rajput/My-Portfolio/web/templates/pages"
)

// AboutHandler serves GET /about
type AboutHandler struct{}

func NewAboutHandler() *AboutHandler {
	return &AboutHandler{}
}

func (h *AboutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pages.About().Render(r.Context(), w)
}
