package main

import "math"

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) Stats {
	stats := Stats{}
	if len(nums) == 0 {
		return stats
	}
	stats.Min = math.MaxInt64
	stats.Max = math.MinInt64
	for _, value := range nums {
		stats.Count++
		stats.Sum += value
		if value > stats.Max {
			stats.Max = value
		}
		if value < stats.Min {
			stats.Min = value
		}
	}
	return stats
}
