package model

// Project holds all metadata and rendered HTML content for a portfolio project.
type Project struct {
	Slug         string
	Title        string
	Tagline      string
	Stack        []string
	GithubURL    string
	Featured     bool
	Problem      string // rendered HTML from markdown
	Architecture string // rendered HTML from markdown
	Decisions    string // rendered HTML from markdown
	Tradeoffs    string // rendered HTML from markdown
	Failures     string // rendered HTML from markdown
	Improvements string // rendered HTML from markdown
}

// AllProjects is the canonical list of projects.
// Markdown content is loaded at startup and merged into each entry.
var AllProjects = []Project{
	{
		Slug:      "flowpay",
		Title:     "FlowPay",
		Tagline:   "Event-driven payment system with idempotent processing and Kafka-backed state machine",
		Stack:     []string{"Go", "gRPC", "Kafka", "PostgreSQL", "Docker"},
		GithubURL: "https://github.com/Bharat1Rajput/flowpay",
		Featured:  true,
	},
	{
		Slug:      "dispatchgo",
		Title:     "DispatchGo",
		Tagline:   "Distributed webhook dispatcher with worker pool, retries, and job lifecycle tracking",
		Stack:     []string{"Go", "RabbitMQ", "PostgreSQL"},
		GithubURL: "https://github.com/Bharat1Rajput/dispatchgo",
		Featured:  true,
	},
	{
		Slug:      "url-shortener",
		Title:     "URL Shortener",
		Tagline:   "Layered architecture with Redis caching, TTL management, and click analytics",
		Stack:     []string{"Go", "Redis", "PostgreSQL"},
		GithubURL: "https://github.com/Bharat1Rajput/url-shortener",
		Featured:  false,
	},
	{
		Slug:      "holdup",
		Title:     "HoldUp",
		Tagline:   "Token bucket rate limiter middleware — thread-safe, O(1), tested with Go race detector",
		Stack:     []string{"Go", "Concurrency", "Rate Limiting"},
		GithubURL: "https://github.com/Bharat1Rajput/holdup",
		Featured:  false,
	},
}
