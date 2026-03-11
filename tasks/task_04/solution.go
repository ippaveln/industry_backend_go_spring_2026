package main

import "math"

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) Stats {
	if len(nums) == 0 {
		return Stats{}
	}

	s := Stats{Min: math.MaxInt64, Max: math.MinInt64}

	for _, num := range nums {
		s.Count++

		if num < s.Min {
			s.Min = num
		}
		
		if num > s.Max {
			s.Max = num
		}

		s.Sum += num
	}

	return s
}
