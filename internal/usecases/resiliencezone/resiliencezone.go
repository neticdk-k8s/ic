package resiliencezone

import "github.com/neticdk/go-common/pkg/types"

func ListResilienceZones() []string {
	return types.AllResilienceZonesString()
}
