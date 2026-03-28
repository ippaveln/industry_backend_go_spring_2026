package main

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) (res Stats) {
	for _, num := range nums {
		res.Count++
		res.Sum += num

		if res.Count == 1 {
			res.Min = num
			res.Max = num
		}
		res.Min = min(res.Min, num)
		res.Max = max(res.Max, num)
	}
	return res
}
