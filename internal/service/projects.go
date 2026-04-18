package service

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Bharat1Rajput/My-Portfolio/internal/model"
	"github.com/Bharat1Rajput/My-Portfolio/renderer"
)

// ProjectService holds all project data in memory after startup.
type ProjectService struct {
	projects map[string]model.Project // keyed by slug
	ordered  []model.Project          // insertion-order slice for listing
}

// NewProjectService reads all project markdown files from contentDir,
// merges rendered HTML into the model.AllProjects metadata, and stores
// everything in memory. Zero disk reads after this returns.
func NewProjectService(contentDir string) (*ProjectService, error) {
	svc := &ProjectService{
		projects: make(map[string]model.Project, len(model.AllProjects)),
	}

	// Build a lookup from slug → index in AllProjects
	meta := make(map[string]model.Project, len(model.AllProjects))
	for _, p := range model.AllProjects {
		meta[p.Slug] = p
	}

	// Load markdown for each project
	for i, p := range model.AllProjects {
		mdPath := filepath.Join(contentDir, p.Slug+".md")
		data, err := os.ReadFile(mdPath)
		if err != nil {
			// Non-fatal: project exists in list but has no markdown detail
			slog.Warn("project markdown not found", "slug", p.Slug, "path", mdPath)
			svc.projects[p.Slug] = p
			svc.ordered = append(svc.ordered, p)
			continue
		}

		sections, err := renderer.ParseProjectMarkdown(data)
		if err != nil {
			return nil, fmt.Errorf("parse markdown for %s: %w", p.Slug, err)
		}

		model.AllProjects[i].Problem = sections.Problem
		model.AllProjects[i].Architecture = sections.Architecture
		model.AllProjects[i].Decisions = sections.Decisions
		model.AllProjects[i].Tradeoffs = sections.Tradeoffs
		model.AllProjects[i].Failures = sections.Failures
		model.AllProjects[i].Improvements = sections.Improvements

		svc.projects[p.Slug] = model.AllProjects[i]
		svc.ordered = append(svc.ordered, model.AllProjects[i])

		slog.Info("loaded project", "slug", p.Slug)
	}

	return svc, nil
}

// All returns projects in definition order.
func (s *ProjectService) All() []model.Project {
	return s.ordered
}

// Featured returns projects marked as Featured.
func (s *ProjectService) Featured() []model.Project {
	var out []model.Project
	for _, p := range s.ordered {
		if p.Featured {
			out = append(out, p)
		}
	}
	return out
}

// BySlug returns a project by slug. Returns false if not found.
func (s *ProjectService) BySlug(slug string) (model.Project, bool) {
	p, ok := s.projects[slug]
	return p, ok
}
