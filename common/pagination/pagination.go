package pagination

import "math"

func From(page int64, pageSize int64) int64 {
	return int64(math.Max(0, float64(pageSize*(page-1))))
}

func Size(pageSize int64) int64 {
	return int64(math.Min(500, math.Max(1, float64(pageSize))))
}

func TotalPages(numberOfResults int64, pageSize int64) int64 {
	pages := math.Ceil(float64(numberOfResults / pageSize))
	if pages == 0.0 {
		return 1
	}
	return int64(pages)
}
