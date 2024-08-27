package controllers

type ListOptions struct {
	Page    int
	PerPage int
}

type PaginatedResource struct {
	Items      []any `json:"items"`
	Page       int   `json:"page"`
	TotalPages int   `json:"total_pages"`
}

type SearchOptions struct {
	ListOptions
	Term        string
	Tags        []string
	Author      string
	Description string
}
