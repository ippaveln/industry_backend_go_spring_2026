package main

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

	stats := Stats{
		Count: 0,
		Sum:   0,
		Min:   nums[0],
		Max:   nums[0],
	}

	for _, v := range nums {
		stats.Count++
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
