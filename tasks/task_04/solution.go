package main

import "math"

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) Stats {
	stat := Stats{}
	if len(nums) == 0 {
		return stat
	}
	stat.Min = math.MaxInt64
	stat.Max = math.MinInt64
	for _, n := range nums {
		stat.Count++
		stat.Sum += n
		stat.Min = min(stat.Min, n)
		stat.Max = max(stat.Max, n)
	}
	return stat
}
