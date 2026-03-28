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

	for _, n := range nums {
		stats.Sum += n
		if n < stats.Min {
			stats.Min = n
		}
		if n > stats.Max {
			stats.Max = n
		}
	}

	return stats
}
