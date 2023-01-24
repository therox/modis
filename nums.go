package modis

import (
	// "constraints"
	"golang.org/x/exp/constraints"
)

func Mean[T constraints.Float | constraints.Integer](nums []T) T {
	if len(nums) == 0 {
		return T(0)
	}
	var sum T
	for _, num := range nums {
		sum += num
	}
	return sum / T(len(nums))
}
