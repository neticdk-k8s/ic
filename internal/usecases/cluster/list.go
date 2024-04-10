package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
)

type Clusters struct {
	// Count is the number of clusters in the collection
	Count int `json:"count"`

	// Clusters is the identification of the clusters in the collection
	Clusters []string `json:"clusters"`

	// Included will container linked resources included here for convenience
	Included []interface{} `json:"@included,omitempty"`
}

type APIClient interface {
	GetClusters(ctx context.Context) (*Clusters, error)
}

type apiClient struct {
	logger    logger.Logger
	client    *http.Client
	serverURL *url.URL
	token     string
}

func NewClient(logger logger.Logger, serverURL string, token string) *apiClient {
	url, _ := url.Parse(serverURL)
	client := &http.Client{}
	return &apiClient{
		logger:    logger,
		client:    client,
		serverURL: url,
		token:     token,
	}
}

func (c *apiClient) GetClusters(ctx context.Context) (*Clusters, error) {
	u := c.serverURL
	u.Path = path.Join(c.serverURL.Path, "clusters")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Info("http", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	clusters := &Clusters{}
	if err := json.NewDecoder(resp.Body).Decode(clusters); err != nil {
		return nil, err
	}

	return clusters, nil
}
