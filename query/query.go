package query

type ListQuery struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}
