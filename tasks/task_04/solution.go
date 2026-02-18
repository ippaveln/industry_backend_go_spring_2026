package main

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) Stats {
	count := len(nums)
	if count == 0 {
		return Stats{Count: count}
	}
	var (
		sum int64 = 0
	)
	min := nums[0]
	max := nums[0]

	for _, num := range nums {
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
		sum += num
	}
	return Stats{
		Count: count,
		Sum:   sum,
		Min:   min,
		Max:   max,
	}
}
