package controllers

type ListOptions struct {
	Page    int
	PerPage int
}

type PaginatedResource struct {
	Items      []any
	Page       int
	TotalPages int
}

type SearchOptions struct {
	ListOptions
	Term        string
	Tags        []string
	Author      string
	Description string
}
