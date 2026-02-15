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
        Count: 1,
        Sum:   nums[0],
        Min:   nums[0],
        Max:   nums[0],
    }
    
    for i := 1; i < len(nums); i++ {
        result.Count++
        result.Sum += nums[i]
        
        if nums[i] < result.Min {
            result.Min = nums[i]
        }
        if nums[i] > result.Max {
            result.Max = nums[i]
        }
    }
    return result
}