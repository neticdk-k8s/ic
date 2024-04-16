package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"

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
	fmt.Fprintln(r.writer, "Base information:")
	table := ui.NewTable(r.writer, []string{})
	data := [][]string{
		{"Name:", r.cluster.Name},
		{"NRN:", r.cluster.NRN},
		{"Provider:", r.cluster.ProviderName},
		{"Description:", r.cluster.Description},
		{"Type:", r.cluster.ClusterType},
		{"Environment:", r.cluster.EnvironmentName},
		{"Resilience Zone:", r.cluster.ResilienceZone},
		{"Infrastructure Provider:", r.cluster.InfrastructureProvider},
		{"Kubernetes Provider:", r.cluster.KubernetesProvider},
		{"Kubernetes Version:", r.cluster.KubernetesVersion},
		{"Client Version:", r.cluster.ClientVersion},
	}
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(false)
	table.AppendBulk(data)
	table.Render()

	if r.cluster.ControlPlaneCapacity != nil {
		fmt.Fprintln(r.writer, "Control Plane Capacity:")
		table = ui.NewTable(r.writer, []string{})
		allocMem, unit := renderAllocMemory(r.cluster.ControlPlaneCapacity.MemoryBytes)
		data = [][]string{
			{"Nodes:", fmt.Sprintf("%d", r.cluster.ControlPlaneCapacity.NodeCount)},
			{"Allocatable CPU (millis):", fmt.Sprintf("%d", r.cluster.ControlPlaneCapacity.CoresMillis)},
			{fmt.Sprintf("Allocatable Memory (%s):", unit), fmt.Sprintf("%.f", allocMem)},
		}
		table.SetTablePadding("  ")
		table.SetNoWhiteSpace(false)
		table.AppendBulk(data)
		table.Render()
	}

	if r.cluster.WorkerNodesCapacity != nil {
		fmt.Fprintln(r.writer, "Worker Nodes Capacity:")
		table = ui.NewTable(r.writer, []string{})
		allocMem, unit := renderAllocMemory(r.cluster.WorkerNodesCapacity.MemoryBytes)
		data = [][]string{
			{"Nodes:", fmt.Sprintf("%d", r.cluster.WorkerNodesCapacity.NodeCount)},
			{"Allocatable CPU (millis):", fmt.Sprintf("%d", r.cluster.WorkerNodesCapacity.CoresMillis)},
			{fmt.Sprintf("Allocatable Memory (%s):", unit), fmt.Sprintf("%.f", allocMem)},
		}
		table.SetTablePadding("  ")
		table.SetNoWhiteSpace(false)
		table.AppendBulk(data)
		table.Render()
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
		headers = []string{"provider", "name", "rz", "version"}
	}
	table := ui.NewTable(r.writer, headers)
	for _, c := range r.clusters.Clusters {
		table.Append(
			[]string{
				c.ProviderName,
				c.Name,
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

func renderAllocMemory(val int64) (newval float64, unit string) {
	gb := math.Pow(1024, 3)
	tb := math.Pow(1024, 4)
	pb := math.Pow(1024, 5)

	newval = float64(val)
	if newval >= pb {
		newval /= pb
		unit = "PB"
	} else if newval >= tb {
		newval /= tb
		unit = "TB"
	} else {
		newval /= gb
		unit = "GB"
	}
	return newval, unit
}
