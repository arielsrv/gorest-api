package paging

type PagedResultDTO[T any] struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Total int `json:"total"`

	Results []T `json:"results"`
}
