package main

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(xs []int64) Stats {
	if len(xs) == 0 {
		return Stats{}
	}

	res := Stats{
		Count: len(xs),
		Sum:   xs[0],
		Min:   xs[0],
		Max:   xs[0],
	}

	for _, v := range xs[1:] {
		res.Sum += v
		if v < res.Min {
			res.Min = v
		}
		if v > res.Max {
			res.Max = v
		}
	}

	return res
}
