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
		return r.renderTable()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *clusterRenderer) renderTable() error {
	table := ui.NewTable(r.writer, []string{})
	data := [][]string{
		{"Name:", r.cluster.Name},
		{"Provider:", r.cluster.ProviderName},
		{"Description:", r.cluster.Description},
		{"Type:", r.cluster.ClusterType},
		{"Environment:", r.cluster.EnvironmentName},
		{"Resilience Zone:", r.cluster.ResilienceZone},
		{"K8S Provider:", r.cluster.KubernetesProvider},
		{"K8S Version:", r.cluster.KubernetesVersion},
	}
	table.SetTablePadding("  ")
	table.AppendBulk(data)
	table.Render()
	return nil
}

func (r *clusterRenderer) renderJSON() error {
	return prettyPrintJSON(r.jsonData, r.writer)
}

type clustersRenderer struct {
	renderer
	clusters *clusterListResponse
}

// NewClustersRenderer creates a new renderer for a list of clusters
func NewClustersRenderer(clusters *clusterListResponse, jsonData []byte, writer io.Writer) *clustersRenderer {
	cr := &clustersRenderer{
		renderer: renderer{
			writer:   writer,
			jsonData: jsonData,
		},
		clusters: clusters,
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
	table := ui.NewTable(r.writer, []string{"provider", "name", "rz", "version"})
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
