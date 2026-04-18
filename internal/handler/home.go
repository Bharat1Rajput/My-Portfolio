package handler

import (
	"net/http"

	"github.com/Bharat1Rajput/My-Portfolio/internal/model"
	"github.com/Bharat1Rajput/My-Portfolio/internal/service"
	"github.com/Bharat1Rajput/My-Portfolio/web/templates/pages"
)

// HomeHandler serves GET /
type HomeHandler struct {
	projects *service.ProjectService
}

func NewHomeHandler(projects *service.ProjectService) *HomeHandler {
	return &HomeHandler{projects: projects}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	featured := h.projects.Featured()
	recentPosts := model.Posts
	if len(recentPosts) > 3 {
		recentPosts = recentPosts[:3]
	}
	pages.Home(featured, recentPosts).Render(r.Context(), w)
}