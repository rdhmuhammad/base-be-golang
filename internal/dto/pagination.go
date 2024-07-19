package dto

type PaginationResponse struct {
	PerPage      int  `json:"perPage"`
	Total        uint `json:"total"`
	CurrentPage  int  `json:"currentPage"`
	PreviousPage int  `json:"previousPage"`
	NextPage     int  `json:"nextPage"`
}

type SummaryStatus struct {
	Label string `json:"label"`
	Type  string `json:"type"`
}

type FilterMapping struct {
	ID    uint   `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

func (r *PaginationResponse) Evaluate() {
	if r.CurrentPage-1 > 0 {
		r.PreviousPage = r.CurrentPage - 1
	}

	if uint(r.CurrentPage*r.PerPage) < r.Total {
		r.NextPage = r.CurrentPage + 1
	}

}
