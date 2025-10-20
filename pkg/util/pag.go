package util

import (
	"math"

	"github.com/vogiaan1904/ticketbottle-order/pkg/paginator"
)

func BuildPagingMeta(page int32, limit int64, total int64) paginator.Paginator {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}
	pageSize := limit
	currentPage := page
	if currentPage == 0 {
		currentPage = 1
	}
	if currentPage > int32(math.Ceil(float64(total)/float64(pageSize))) {
		currentPage = 1
	}

	var count int64
	if limit > total {
		count = total
	} else {
		if int64(currentPage)*limit > int64(total) {
			count = total - (int64(currentPage)-1)*limit
		} else {
			count = limit
		}
	}
	return paginator.Paginator{
		Total:    total,
		Count:    count,
		PageSize: pageSize,
		Page:     currentPage,
	}
}
