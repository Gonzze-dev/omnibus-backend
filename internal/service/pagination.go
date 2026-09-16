package service

import "strings"

// normalizePagination aplica los defaults y topes de los parámetros de paginación.
// Devuelve page (>=1), limit (entre 1 y maxLimit) y order normalizado a "ASC" o "DESC".
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

// pageLinks calcula las páginas anterior y siguiente. prev es nil en la primera
// página y next es nil en la última.
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
