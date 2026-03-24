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

	stats := Stats{
		Count: len(xs),
		Min:   xs[0],
		Max:   xs[0],
	}

	for _, x := range xs {
		stats.Sum += x
		if x < stats.Min {
			stats.Min = x
		}
		if x > stats.Max {
			stats.Max = x
		}
	}

	return stats
}