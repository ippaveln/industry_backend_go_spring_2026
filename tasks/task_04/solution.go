package main

type Stats struct {
	Count int64
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) Stats {
	if len(nums) == 0 {
		return Stats{
			0, 0, 0, 0,
		}
	}

	var currStats = Stats{
		0, 0, nums[0], nums[0],
	}

	for _, el := range nums {
		currStats.Count++
		currStats.Sum += el

		if el < currStats.Min {
			currStats.Min = el
		}
		if el > currStats.Max {
			currStats.Max = el
		}
	}
	return currStats
}
