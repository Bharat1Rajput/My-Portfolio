package handler

import (
	"net/http"

	"github.com/Bharat1Rajput/My-Portfolio/internal/model"
	"github.com/Bharat1Rajput/My-Portfolio/web/templates/pages"
)

// BlogHandler serves GET /blog
type BlogHandler struct{}

func NewBlogHandler() *BlogHandler {
	return &BlogHandler{}
}

func (h *BlogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pages.Blog(model.Posts).Render(r.Context(), w)
}
