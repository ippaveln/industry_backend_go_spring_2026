package main

import "math"

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) Stats {
	var stats Stats

	if len(nums) == 0 {
		return stats
	}

	stats.Count = len(nums)

	max := int64(math.MinInt64)
	min := int64(math.MaxInt64)

	for i := 0; i < len(nums); i++ {
		stats.Sum += nums[i]

		if max < nums[i] {
			max = nums[i]
		}

		if min > nums[i] {
			min = nums[i]
		}
	}

	stats.Max = max
	stats.Min = min

	return stats
}
