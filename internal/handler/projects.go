package handler

import (
	"net/http"

	"github.com/Bharat1Rajput/My-Portfolio/internal/service"
	"github.com/Bharat1Rajput/My-Portfolio/web/templates/pages"
	"github.com/go-chi/chi/v5"
)

// ProjectsHandler serves GET /projects and GET /projects/{slug}
type ProjectsHandler struct {
	projects *service.ProjectService
}

func NewProjectsHandler(projects *service.ProjectService) *ProjectsHandler {
	return &ProjectsHandler{projects: projects}
}

// List serves GET /projects
func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	all := h.projects.All()
	pages.Projects(all).Render(r.Context(), w)
}

// Detail serves GET /projects/{slug}
func (h *ProjectsHandler) Detail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	project, ok := h.projects.BySlug(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	pages.ProjectDetail(project).Render(r.Context(), w)
}
