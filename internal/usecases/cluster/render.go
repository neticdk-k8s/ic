package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/neticdk-k8s/ic/internal/ui"
)

type Renderer interface {
	Render(format string) error
}

type renderer struct {
	jsonData []byte
	writer   io.Writer
}

type clusterRenderer struct {
	renderer
	cluster *clusterResponse
}

// NewClusterRenderer creates a new renderer of a single cluster
func NewClusterRenderer(cluster *clusterResponse, jsonData []byte, writer io.Writer) *clusterRenderer {
	cr := &clusterRenderer{
		renderer: renderer{
			jsonData: jsonData,
			writer:   writer,
		},
		cluster: cluster,
	}
	return cr
}

// Render renders the cluster
func (r *clusterRenderer) Render(format string) error {
	switch format {
	case "json":
		return r.renderJSON()
	case "text", "table":
		return r.renderText()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *clusterRenderer) renderText() error {
	data := [][]string{
		{"ID:", fmt.Sprintf("%s.%s", r.cluster.Name, r.cluster.ProviderName)},
		{"Name:", r.cluster.Name},
		{"Provider:", r.cluster.ProviderName},
		{"NRN:", r.cluster.NRN},
		{"Description:", r.cluster.Description},
		{"Type:", r.cluster.ClusterType},
		{"Environment:", r.cluster.EnvironmentName},
		{"Resilience Zone:", r.cluster.ResilienceZone},
		{"Infrastructure Provider:", r.cluster.InfrastructureProvider},
		{"Kubernetes Provider:", r.cluster.KubernetesProvider},
		{"Kubernetes Version:", r.cluster.KubernetesVersion},
		{"Client Version:", r.cluster.ClientVersion},
	}
	ui.RenderKVTable(r.writer, "Base Information", data)

	if r.cluster.ControlPlaneCapacity != nil {
		allocMem, unit := bytesToBinarySI(r.cluster.ControlPlaneCapacity.MemoryBytes)
		data = [][]string{
			{"Nodes:", fmt.Sprintf("%d", r.cluster.ControlPlaneCapacity.NodeCount)},
			{"Allocatable CPU:", fmt.Sprintf("%dm", r.cluster.ControlPlaneCapacity.CoresMillis)},
			{"Allocatable Memory:", fmt.Sprintf("%.f%s", allocMem, unit)},
		}
		ui.RenderKVTable(r.writer, "Control Plane Capacity", data)
	}

	if r.cluster.WorkerNodesCapacity != nil {
		allocMem, unit := bytesToBinarySI(r.cluster.WorkerNodesCapacity.MemoryBytes)
		data = [][]string{
			{"Nodes:", fmt.Sprintf("%d", r.cluster.WorkerNodesCapacity.NodeCount)},
			{"Allocatable CPU:", fmt.Sprintf("%dm", r.cluster.WorkerNodesCapacity.CoresMillis)},
			{"Allocatable Memory:", fmt.Sprintf("%.f%s", allocMem, unit)},
		}
		ui.RenderKVTable(r.writer, "Worker Nodes Capacity", data)
	}

	return nil
}

func (r *clusterRenderer) renderJSON() error {
	return prettyPrintJSON(r.jsonData, r.writer)
}

type clustersRenderer struct {
	renderer
	noHeaders bool
	clusters  *clusterListResponse
}

// NewClustersRenderer creates a new renderer for a list of clusters
func NewClustersRenderer(clusters *clusterListResponse, jsonData []byte, writer io.Writer, noHeaders bool) *clustersRenderer {
	cr := &clustersRenderer{
		renderer: renderer{
			writer:   writer,
			jsonData: jsonData,
		},
		noHeaders: noHeaders,
		clusters:  clusters,
	}
	return cr
}

// Render renders the cluster list
func (r *clustersRenderer) Render(format string) error {
	switch format {
	case "json":
		return r.renderJSON()
	case "text", "table":
		return r.renderTable()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *clustersRenderer) renderTable() error {
	var headers []string
	if !r.noHeaders {
		headers = []string{"provider", "id", "rz", "version"}
	}
	table := ui.NewTable(r.writer, headers)
	for _, c := range r.clusters.Clusters {
		table.Append(
			[]string{
				c.ProviderName,
				fmt.Sprintf("%s.%s", c.Name, c.ProviderName),
				c.ResilienceZone,
				c.KubernetesVersion,
			},
		)
	}
	table.Render()
	return nil
}

func (r *clustersRenderer) renderJSON() error {
	return prettyPrintJSON(r.jsonData, r.writer)
}

func prettyPrintJSON(body []byte, writer io.Writer) error {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(writer, prettyJSON.String())
	return nil
}

func bytesToBinarySI(bytes int64) (float64, string) {
	const (
		kibi float64 = 1024
		mebi float64 = 1048576
		gibi float64 = 1073741824
		tebi float64 = 1099511627776
		pebi float64 = 1125899906842624
	)

	b := float64(bytes)
	if b >= pebi {
		return b / pebi, "PiB"
	} else if b >= tebi {
		return b / tebi, "TiB"
	} else if b >= gibi {
		return b / gibi, "GiB"
	} else if b >= mebi {
		return b / mebi, "MiB"
	} else if b >= kibi {
		return b / kibi, "KiB"
	}
	return b, "B"
}
