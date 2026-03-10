package main

type Stats struct {
	Count int64
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) Stats {
	if len(nums) == 0 {
		return Stats{}
	}
	var count, sum int64
	min, max := nums[0], nums[0]
	for _, n := range nums {
		count += 1 //Можно было сделать int64(len(nums)), но это уже не счетчик какой-то получается
		sum += n
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return Stats{count, sum, min, max}
}
