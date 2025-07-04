package responses

type PaginationResponse struct {
	Data         any   `json:"data"`
	TotalCount   int64 `json:"totalCount"`
	TotalPage    int64 `json:"totalPage"`
	CurrentPage  int64 `json:"currentPage"`
	LastPage     int64 `json:"lastPage"`
	PerPage      int64 `json:"perPage"`
	NextPage     int64 `json:"nextPage"`
	PreviousPage int64 `json:"previousPage"`
}
