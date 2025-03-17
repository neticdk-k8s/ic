package component

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/neticdk-k8s/ic/internal/apiclient"
)

type componentResponse struct {
	ID              string                   `json:"id,omitempty"`
	Name            string                   `json:"name,omitempty"`
	Namespace       string                   `json:"namespace,omitempty"`
	Description     string                   `json:"description,omitempty"`
	ComponentType   string                   `json:"component_type,omitempty"`
	Source          string                   `json:"source,omitempty"`
	ResilienceZones []resilienceZoneResponse `json:"resilience_zones,omitempty"`
	Clusters        []string                 `json:"clusters,omitempty"`
}

type resilienceZoneResponse struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type componentListResponse struct {
	Components []componentResponse `json:"components,omitempty"`
}

// ComponentList is a list of components
type ComponentList struct { //nolint
	// Included is a list of included items
	Included []map[string]any
	// Components is a list of components
	Components []string
}

func (cl *ComponentList) ToResponse() *componentListResponse {
	clr := &componentListResponse{
		Components: make([]componentResponse, 0),
	}
	includeMap := make(map[string]any)
	for _, i := range cl.Included {
		if v, ok := mapValAs[string](i, "@id"); ok {
			includeMap[v] = i
		}
	}
	for _, i := range cl.Included {
		if v, ok := mapValAs[string](i, "@type"); ok {
			if v != "Component" {
				continue
			}
		}
		cr := componentResponse{}
		cr.Name, _ = mapValAs[string](i, "name")
		cr.Namespace, _ = mapValAs[string](i, "namespace")
		cr.ComponentType, _ = mapValAs[string](i, "component_type")
		cr.ID = fmt.Sprintf("%s.%s", cr.Name, cr.Namespace)
		cr.Description, _ = mapValAs[string](i, "description")
		cr.Source, _ = mapValAs[string](i, "source")
		cr.Clusters = make([]string, 0)
		if clusters, ok := mapValAs[[]any](i, "clusters"); ok {
			for _, c := range clusters {
				if cluster, ok := c.(map[string]any); ok {
					if clusterID, ok := mapValAs[string](cluster, "cluster_id"); ok {
						cr.Clusters = append(cr.Clusters, clusterID)
					}
				}
			}
		}
		cr.ResilienceZones = make([]resilienceZoneResponse, 0)
		if resilienceZones, ok := mapValAs[[]any](i, "resilience_zones"); ok {
			for _, rz := range resilienceZones {
				res := resilienceZoneResponse{}
				if rezilienceZone, ok := rz.(map[string]any); ok {
					if rzName, ok := mapValAs[string](rezilienceZone, "name"); ok {
						res.Name = rzName
					}
					if rzVersion, ok := mapValAs[string](rezilienceZone, "version"); ok {
						res.Version = rzVersion
					}
					cr.ResilienceZones = append(cr.ResilienceZones, res)
				}
			}
		}
		clr.Components = append(clr.Components, cr)
	}
	return clr
}

func (cl *ComponentList) MarshalJSON() ([]byte, error) {
	return json.Marshal(cl.ToResponse())
}

// ListComponentsInput is the input given to ListComponents()
type ListComponentsInput struct {
	// Logger is a logger
	Logger *slog.Logger
	// APIClient is the inventory server API client used to make requests
	APIClient apiclient.ClientWithResponsesInterface
}

// ListComponentResults is the result of ListComponents
type ListComponentResults struct {
	ComponentListResponse *componentListResponse
	JSONResponse          []byte
	Problem               *apiclient.Problem
}

// ListComponents returns a non-paginated list of components
func ListComponents(ctx context.Context, in ListComponentsInput) (*ListComponentResults, error) {
	cl := &ComponentList{}
	problem, err := listComponents(ctx, &in, cl)
	if err != nil {
		return nil, fmt.Errorf("listComponents: %w", err)
	}
	if problem != nil {
		return &ListComponentResults{nil, nil, problem}, nil
	}
	jsonData, err := cl.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshaling component list: %w", err)
	}
	return &ListComponentResults{cl.ToResponse(), jsonData, nil}, nil
}

func listComponents(ctx context.Context, in *ListComponentsInput, componentList *ComponentList) (*apiclient.Problem, error) { //nolint
	response, err := in.APIClient.ListComponentsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading components: %w", err)
	}
	in.Logger.DebugContext(ctx, "listComponents", logStatus(response.HTTPResponse)...)
	switch response.StatusCode() {
	case http.StatusOK:
	case http.StatusBadRequest:
		return response.ApplicationproblemJSON400, nil
	case http.StatusInternalServerError:
		return response.ApplicationproblemJSON500, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}
	if response.ApplicationldJSONDefault.Components != nil {
		componentList.Components = append(componentList.Components, *response.ApplicationldJSONDefault.Components...)
	}
	if response.ApplicationldJSONDefault.Included != nil {
		componentList.Included = append(componentList.Included, *response.ApplicationldJSONDefault.Included...)
	}
	return nil, nil
}

// GetComponentInput is the input used by GetComponent()
type GetComponentInput struct {
	Logger    *slog.Logger
	APIClient apiclient.ClientWithResponsesInterface
}

// GetComponentResult is the result of GetComponent
type GetComponentResult struct {
	ComponentResponse *componentResponse
	JSONResponse      []byte
	Problem           *apiclient.Problem
}

// GetComponent returns information abuot a component
func GetComponent(ctx context.Context, namespace, name string, in GetComponentInput) (*GetComponentResult, error) {
	response, err := in.APIClient.GetComponentWithResponse(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("getComponent: %w", err)
	}
	in.Logger.DebugContext(ctx, "apiclient", logStatus(response.HTTPResponse)...)
	switch response.StatusCode() {
	case http.StatusOK:
	case http.StatusNotFound:
		return &GetComponentResult{nil, nil, response.ApplicationproblemJSON404}, nil
	case http.StatusInternalServerError:
		return &GetComponentResult{nil, nil, response.ApplicationproblemJSON500}, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}

	component := toComponentResponse(response.ApplicationldJSONDefault)

	jsonData, err := json.Marshal(component)
	if err != nil {
		return nil, fmt.Errorf("marshaling component: %w", err)
	}

	return &GetComponentResult{component, jsonData, nil}, nil
}

func toComponentResponse(component *apiclient.Component) *componentResponse {
	includeMap := make(map[string]any)
	if component.Included != nil {
		for _, i := range *component.Included {
			if v, ok := mapValAs[string](i, "@id"); ok {
				includeMap[v] = i
			}
		}
	}
	cr := &componentResponse{}
	cr.Name = nilStr(component.Name)
	cr.ComponentType = nilStr(component.ComponentType)
	cr.Namespace = nilStr(component.Namespace)
	cr.Source = nilStr(component.Source)
	cr.Description = nilStr(component.Description)
	cr.Clusters = make([]string, 0)
	if component.Clusters != nil {
		for _, c := range *component.Clusters {
			cr.Clusters = append(cr.Clusters, nilStr(c.ClusterId))
		}
	}
	cr.ResilienceZones = make([]resilienceZoneResponse, 0)
	if component.ResilienceZones != nil {
		for _, rz := range *component.ResilienceZones {
			cr.ResilienceZones = append(cr.ResilienceZones, resilienceZoneResponse{
				Name:    nilStr(rz.Name),
				Version: nilStr(rz.Version),
			})
		}
	}
	return cr
}

func nilStr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func mapValAs[T any](haystak map[string]any, needle string) (T, bool) {
	if v, ok := haystak[needle]; ok {
		if v2, ok := v.(T); ok {
			return v2, true
		}
	}
	var zero T
	return zero, false
}

func logStatus(r *http.Response) []any {
	return []any{
		slog.Int("status", r.StatusCode),
		slog.String("content-type", r.Header.Get("Content-Type")),
	}
}
