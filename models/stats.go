package models

type StatsResponse struct {
	TotalProjects   int               `json:"total_projects"`
	TotalErrorGroups int              `json:"total_error_groups"`
	TotalErrors      int              `json:"total_errors"`
	ByStatus         map[string]int   `json:"by_status"`
	ErrorsLast24h    int              `json:"errors_last_24h"`
}
