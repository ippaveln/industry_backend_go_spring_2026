package main

import "math"

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) (res Stats) {	
	if len(nums) == 0 {
		return res
	}
	
	res.Count = len(nums)
	res.Min = int64(math.MaxInt64)
	res.Max = int64(math.MinInt64)

	for i := range nums {
		res.Sum += nums[i]
		if nums[i] < res.Min {
			res.Min = nums[i]
		}
		if nums[i] > res.Max {
			res.Max = nums[i]
		}
	}

	return res
}
