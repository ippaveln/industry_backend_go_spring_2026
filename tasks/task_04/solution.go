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

	var s Stats
	s.Count = len(nums)
	s.Sum = 0
	s.Min = nums[0]
	s.Max = nums[0]

	for _, v := range nums {
		s.Sum += v
		if v < s.Min {
			s.Min = v
		}
		if v > s.Max {
			s.Max = v
		}
	}

	return s
}
