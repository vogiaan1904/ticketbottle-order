package paginator

import "math"

const (
	defaultPage  = 1
	defaultLimit = 15
)

// PaginatorQuery is a struct that contains the page and limit of a request.
type PaginatorQuery struct {
	Page        int32 `json:"page" form:"page"`
	Limit       int64 `json:"limit" form:"limit"`
	ShiftOffset int64
}

// Adjust adjusts the paginator's page and limit to the default values if they are invalid.
func (p *PaginatorQuery) Adjust() {
	if p.Page < 1 {
		p.Page = defaultPage
	}

	if p.Limit < 1 {
		p.Limit = defaultLimit
	}
}

// Offset returns the offset of the paginator.
func (p *PaginatorQuery) Offset() int64 {
	offset := int64(p.Page-1)*p.Limit - p.ShiftOffset
	if offset < 0 {
		return 0
	}
	return offset
}

type Paginator struct {
	Total    int64
	Count    int64
	PageSize int64
	Page     int32
}

// TotalPages returns the total pages of the paginator.
func (p Paginator) LastPage() int32 {
	if p.Total == 0 {
		return 0
	}

	return int32(math.Ceil(float64(p.Total) / float64(p.PageSize)))
}

// ToResponse converts the paginator to a response.
func (p Paginator) ToResponse() PaginatorResponse {
	return PaginatorResponse{
		Total:       p.Total,
		Count:       p.Count,
		PageSize:    p.PageSize,
		Page:        p.Page,
		LastPage:    p.LastPage(),
		HasNext:     p.Page < p.LastPage(),
		HasPrevious: p.Page > 1,
	}
}

// PaginatorResponse is a struct that contains the response of a paginator.
type PaginatorResponse struct {
	Total       int64 `json:"total"`
	Page        int32 `json:"page"`
	PageSize    int64 `json:"page_size"`
	Count       int64 `json:"count"`
	LastPage    int32 `json:"last_page"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}
