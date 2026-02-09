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

	var sum int64
	minCounter := nums[0]
	maxCounter := nums[0]

	for _, value := range nums {
		sum += value

		if value < minCounter {
			minCounter = value
		}

		if value > maxCounter {
			maxCounter = value
		}
	}

	return Stats{
		Count: len(nums),
		Sum:   sum,
		Min:   minCounter,
		Max:   maxCounter,
	}
}
