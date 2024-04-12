package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/apiclient"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/ui"
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
	cluster *apiclient.Cluster
}

func NewClusterRenderer(cluster *apiclient.Cluster, jsonData []byte, writer io.Writer) Renderer {
	cr := &clusterRenderer{
		renderer: renderer{
			jsonData: jsonData,
			writer:   writer,
		},
		cluster: cluster,
	}
	return cr
}

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
		{"Name", ":", *r.cluster.Name},
		{"Provider", ":", *r.cluster.Provider},
		{"Description", ":", *r.cluster.Description},
		{"Type", ":", *r.cluster.ClusterType},
		{"Environment", ":", *r.cluster.EnvironmentName},
		{"Resilience Zone", ":", rzFromURL(*r.cluster.ResilienceZone)},
		{"K8S Provider", ":", *r.cluster.KubernetesProvider},
		{"K8S Version", ":", *r.cluster.KubernetesVersion.Version},
	}
	table.SetTablePadding(" ")
	table.AppendBulk(data)
	table.Render()
	return nil
}

func (r *clusterRenderer) renderJSON() error {
	return prettyPrintJSON(r.jsonData, r.writer)
}

type clustersRenderer struct {
	renderer
	clusters *ClusterList
}

func NewClustersRenderer(clusters *ClusterList, jsonData []byte, writer io.Writer) Renderer {
	cr := &clustersRenderer{
		renderer: renderer{
			writer:   writer,
			jsonData: jsonData,
		},
		clusters: clusters,
	}
	return cr
}

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
	for _, i := range r.clusters.Included {
		if i["@type"] != "Cluster" {
			continue
		}
		table.Append(
			[]string{
				i["provider"].(string),
				i["name"].(string),
				rzFromURL(i["resilienceZone"].(string)),
				i["kubernetesVersion"].(map[string]interface{})["version"].(string),
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

func rzFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return parts[0]
}
