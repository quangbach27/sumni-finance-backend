package common

type Pagination[T any] struct {
	Items      []T
	TotalCount int
	Page       int
	PageSize   int
}
