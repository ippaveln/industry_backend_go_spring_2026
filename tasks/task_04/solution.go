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
	ans := Stats{Count: 0, Sum: 0, Min: nums[0], Max: nums[0]}
	for _, num := range nums {
		ans.Count++
		ans.Sum += num
		ans.Min = min(ans.Min, num)
		ans.Max = max(ans.Max, num)
	}
	return ans
}
