package cluster

import (
	"context"
	"fmt"
	"net/http"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/apiclient"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
)

type ListClustersInput struct {
	Logger    logger.Logger
	APIClient *apiclient.ClientWithResponses
}

func ListClusters(ctx context.Context, in ListClustersInput) (*apiclient.ListClustersResponse, error) {
	clusters, err := in.APIClient.ListClustersWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading clusters: %w", err)
	}
	in.Logger.Debug("apiclient", "status", clusters.StatusCode(), "content-type", clusters.HTTPResponse.Header.Get("Content-Type"))
	if clusters.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", clusters.StatusCode())
	}
	return clusters, nil
}

type GetClusterInput struct {
	Logger    logger.Logger
	APIClient *apiclient.ClientWithResponses
}

func GetCluster(ctx context.Context, clusterID string, in GetClusterInput) (*apiclient.GetClusterResponse, error) {
	cluster, err := in.APIClient.GetClusterWithResponse(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("getting cluster: %w", err)
	}
	in.Logger.Debug("apiclient", "status", cluster.StatusCode(), "content-type", cluster.HTTPResponse.Header.Get("Content-Type"))

	if cluster.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", cluster.StatusCode())
	}
	return cluster, nil
}
