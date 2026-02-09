package main

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) Stats {

	var min int64 = 10000000000
	var max int64 = -1000000000
	var sum int64 = 0
	for i := 0; i < len(nums); i += 1 {
		sum += nums[i]
		if max < nums[i] {
			max = nums[i]
		}
		if min > nums[i] {
			min = nums[i]
		}
	}

	var stats = Stats{}
	if len(nums) != 0 {
		stats.Count = len(nums)
		stats.Sum = sum
		stats.Max = max
		stats.Min = min
	}
	return stats
}
