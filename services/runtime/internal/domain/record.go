package domain

type Record map[string]any

type ListResult struct {
	Items []Record `json:"items"`
	Total int      `json:"total"`
}
