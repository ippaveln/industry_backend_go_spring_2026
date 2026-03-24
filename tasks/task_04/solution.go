package main

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
	stats.Min, stats.Max = nums[0], nums[0]
	for _, v := range nums {
		stats.Count = len(nums)
		stats.Sum += v
		if v < stats.Min {
			stats.Min = v
		}
		if v > stats.Max {
			stats.Max = v
		}
	}
	return stats
}
