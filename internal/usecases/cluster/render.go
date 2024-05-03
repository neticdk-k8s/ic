package cluster

import (
	"fmt"
	"io"

	"github.com/neticdk-k8s/ic/internal/render"
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
		{"Partition:", r.cluster.Partition},
		{"Region:", r.cluster.Region},
		{"Environment:", r.cluster.EnvironmentName},
		{"Resilience Zone:", r.cluster.ResilienceZone},
		{"Infrastructure Provider:", r.cluster.InfrastructureProvider},
		{"Kubernetes Provider:", r.cluster.KubernetesProvider},
		{"Kubernetes Version:", r.cluster.KubernetesVersion},
		{"Client Version:", r.cluster.ClientVersion},
	}
	ui.RenderKVTable(r.writer, "Base Information", data)

	if r.cluster.ControlPlaneCapacity != nil {
		allocMem, unit := render.BytesToBinarySI(r.cluster.ControlPlaneCapacity.MemoryBytes)
		data = [][]string{
			{"Nodes:", fmt.Sprintf("%d", r.cluster.ControlPlaneCapacity.NodeCount)},
			{"Allocatable CPU:", fmt.Sprintf("%dm", r.cluster.ControlPlaneCapacity.CoresMillis)},
			{"Allocatable Memory:", fmt.Sprintf("%.f%s", allocMem, unit)},
		}
		ui.RenderKVTable(r.writer, "Control Plane Capacity", data)
	}

	if r.cluster.WorkerNodesCapacity != nil {
		allocMem, unit := render.BytesToBinarySI(r.cluster.WorkerNodesCapacity.MemoryBytes)
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
	return render.PrettyPrintJSON(r.jsonData, r.writer)
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
	return render.PrettyPrintJSON(r.jsonData, r.writer)
}

type clusterNodesRenderer struct {
	renderer
	noHeaders bool
	nodes     *clusterNodesListResponse
}

// NewClusterNodesRenderer creates a new renderer for a list of cluster nodes
func NewClusterNodesRenderer(nodes *clusterNodesListResponse, jsonData []byte, writer io.Writer, noHeaders bool) *clusterNodesRenderer {
	cnr := &clusterNodesRenderer{
		renderer: renderer{
			writer:   writer,
			jsonData: jsonData,
		},
		noHeaders: noHeaders,
		nodes:     nodes,
	}
	return cnr
}

// Render renders the cluster node list
func (r *clusterNodesRenderer) Render(format string) error {
	switch format {
	case "json":
		return r.renderJSON()
	case "text", "table":
		return r.renderTable()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *clusterNodesRenderer) renderTable() error {
	var headers []string
	if !r.noHeaders {
		headers = []string{"name", "cp", "kubelet", "cpu (alloc)", "mem (alloc)", "cpu (cap)", "mem (cap)"}
	}
	table := ui.NewTable(r.writer, headers)
	for _, n := range r.nodes.Nodes {
		allocMem, allocMemUnit := render.BytesToBinarySI(int64(n.AllocatableMemoryBytes))
		capMem, capMemUnit := render.BytesToBinarySI(int64(n.CapacityMemoryBytes))
		table.Append(
			[]string{
				n.Name,
				fmt.Sprintf("%t", n.IsControlPlane),
				n.KubeletVersion,
				fmt.Sprintf("%dm", int64(n.AllocatableCPUMillis)),
				fmt.Sprintf("%.f%s", allocMem, allocMemUnit),
				fmt.Sprintf("%dm", int64(n.CapacityCPUMillis)),
				fmt.Sprintf("%.f%s", capMem, capMemUnit),
			},
		)
	}
	table.Render()
	return nil
}

func (r *clusterNodesRenderer) renderJSON() error {
	return render.PrettyPrintJSON(r.jsonData, r.writer)
}
