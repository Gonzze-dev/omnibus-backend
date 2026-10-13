package service

import "strings"

// normalizePagination applies the defaults and caps of the pagination parameters.
// Returns page (>=1), limit (between 1 and maxLimit) and order normalized to "ASC" or "DESC".
func normalizePagination(page, limit int, order string, defaultLimit, maxLimit int) (int, int, string) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	normalized := strings.ToUpper(strings.TrimSpace(order))
	if normalized != "ASC" {
		normalized = "DESC"
	}
	return page, limit, normalized
}

// pageLinks computes the previous and next pages. prev is nil on the first
// page and next is nil on the last one.
func pageLinks(total int64, page, limit int) (next, prev *int) {
	lastPage := int((total + int64(limit) - 1) / int64(limit))
	if lastPage < 1 {
		lastPage = 1
	}
	if page > 1 {
		p := page - 1
		prev = &p
	}
	if page < lastPage {
		n := page + 1
		next = &n
	}
	return next, prev
}
