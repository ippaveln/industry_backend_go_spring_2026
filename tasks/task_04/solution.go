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
	res := Stats{
		Count: 1,
		Sum:   nums[0],
		Min:   nums[0],
		Max:   nums[0],
	}
	for i := 1; i < len(nums); i++ {
		v := nums[i]
		res.Count++
		res.Sum += v
		res.Min = min(res.Min, v)
		res.Max = max(res.Max, v)
	}
	return res
}
