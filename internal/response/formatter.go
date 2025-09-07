package response

type ListResponse[T any] struct {
	Data  []T
	Count int
}
