package partition

import "github.com/neticdk/go-common/pkg/types"

func ListPartitions() []string {
	return types.AllPartitionsString()
}
