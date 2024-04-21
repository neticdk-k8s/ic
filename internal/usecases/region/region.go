package region

import (
	"fmt"

	"github.com/neticdk/go-common/pkg/types"
)

func ListRegions() []string {
	return types.AllRegionsString()
}

func ListRegionsForPartition(p string) (partitions []string, err error) {
	if p == "" {
		return
	}
	part, ok := types.ParsePartition(p)
	if !ok {
		return partitions, fmt.Errorf(`invalid partition: %s`, p)
	}
	regions := types.PartitionRegions(part)
	for _, r := range regions {
		partitions = append(partitions, r.String())
	}
	return
}
