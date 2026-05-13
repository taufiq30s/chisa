package music

import (
	"github.com/disgoorg/disgolink/v3/lavalink"
)

const (
	maxCPUThreshold          = 80.0 // Maximum CPU usage threshold (%)
	maxMemoryThreshold       = 90.0 // Maximum memory usage threshold (%)
	maxFrameDeficitThreshold = 200  // Maximum frame deficit threshold (ms)
	minPlayersThreshold      = 1    // Minimum players threshold
)

func checkNodeHealth(nodeStats *lavalink.Stats) bool {
	if nodeStats.Memory.Allocated == 0 {
		return false
	}
	memUsagePct := (float64(nodeStats.Memory.Used) / float64(nodeStats.Memory.Allocated)) * 100
	return nodeStats.CPU.SystemLoad < maxCPUThreshold &&
		memUsagePct < maxMemoryThreshold &&
		nodeStats.Players > minPlayersThreshold
}
