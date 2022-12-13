package modis

func Mean(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	var sum float64
	for _, num := range nums {
		sum += num
	}
	return sum / float64(len(nums))
}
