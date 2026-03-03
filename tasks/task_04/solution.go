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
		Count: len(nums),
		Min:   nums[0],
		Max:   nums[0],
	}

	for _, val := range nums {
		stats.Sum += val
		if val < stats.Min {
			stats.Min = val
		} else if val > stats.Max {
			stats.Max = val
		}
	}

	return stats
}
