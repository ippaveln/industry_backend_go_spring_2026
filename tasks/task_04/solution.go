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

	result := Stats{
		Count: len(nums),
		Sum:   nums[0],
		Max:   nums[0],
		Min:   nums[0],
	}

	for _, v := range nums[1:] {
		result.Sum += v
		if v < result.Min {
			result.Min = v
		}
		if v > result.Max {
			result.Max = v
		}
	}

	return result
}
