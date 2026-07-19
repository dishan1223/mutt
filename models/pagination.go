package models

import "strconv"

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

func ParsePagination(pageStr, limitStr string) (int, int) {
	page, limit := 1, 20
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	return page, limit
}

func NewPagination(page, limit int, totalCount int64) Pagination {
	if limit <= 0 {
		limit = 20
	}
	totalPages := int(totalCount) / limit
	if int(totalCount)%limit > 0 {
		totalPages++
	}
	return Pagination{
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}

type PaginatedProjects struct {
	Projects   []ProjectResponse `json:"projects"`
	Pagination Pagination        `json:"pagination"`
}

type PaginatedErrorGroups struct {
	ErrorGroups []ErrorGroupResponse `json:"error_groups"`
	Pagination  Pagination           `json:"pagination"`
}
