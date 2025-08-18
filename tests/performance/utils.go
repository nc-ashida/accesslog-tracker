package performance

import (
	"sort"
	"time"
)

// CalculatePercentile は指定されたパーセンタイルの値を計算します
func CalculatePercentile(latencies []time.Duration, percentile int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// ソート
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	index := int(float64(len(latencies)-1) * float64(percentile) / 100.0)
	return latencies[index]
}
