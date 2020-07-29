package customquery

// Result defines a struct for custom query result.
type Result struct {
	Limit      int         `json:"limit"`
	Page       int         `json:"page"`
	TotalPage  int         `json:"total_page"`
	TotalCount int         `json:"total_count"`
	Data       interface{} `json:"-"`
}
