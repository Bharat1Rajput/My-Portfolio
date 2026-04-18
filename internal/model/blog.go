package model

// BlogLink represents a post published on Hashnode.
// Update this slice manually when you publish a new post.
type BlogLink struct {
	Title   string
	Summary string
	URL     string
	Date    string
	Tags    []string
}

// Posts is the hardcoded list of Hashnode blog posts.
var Posts = []BlogLink{

	{
		Title:   "Why I Chose Token Bucket for HoldUp (And Why the Others Didn't Make the Cut)",
		Summary: "A deep dive into how I chose the Token Bucket algorithm for HoldUp — not as a textbook exercise, but as a real trade-off between burst tolerance, fairness, and simplicity.",
		URL:     "https://bharatsingh.hashnode.dev/why-i-chose-token-bucket-for-holdup",
		Date:    "2025-04-18",
		Tags:    []string{"go", "ratelimiting", "backend", "production", "algorithms"},
	},
	{
		Title:   "How First Principles Finally Made Distributed Systems Click for Me",
		Summary: "I used to know the terms but never really got distributed systems. This is the mental shift — network failures, partial failure, idempotency — that changed how I think about building them.",
		URL:     "https://bharatsingh.hashnode.dev/distributed-systems",
		Date:    "2025-04-18",
		Tags:    []string{"distributed-systems", "first-principles", "backend", "beginners"},
	},
	{
		Title:   "When Things Break: Reliability and Failure Handling in Distributed Systems",
		Summary: "Retries, idempotency, timeouts, circuit breakers, and rate limiting — not as isolated patterns, but how they connect to build systems that survive production.",
		URL:     "https://bharatsingh.hashnode.dev/best-production-practices",
		Date:    "2025-04-18",
		Tags:    []string{"distributedsystems", "backend", "go", "reliability", "production"},
	},
}
