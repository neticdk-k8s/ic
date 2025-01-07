package component

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/neticdk-k8s/ic/internal/apiclient"
	"github.com/neticdk-k8s/ic/internal/logger"
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

type ComponentList struct {
	Included   []map[string]interface{}
	Components []string
}

func (cl *ComponentList) ToResponse() *componentListResponse {
	clr := &componentListResponse{
		Components: make([]componentResponse, 0),
	}
	includeMap := make(map[string]interface{})
	for _, i := range cl.Included {
		includeMap[i["@id"].(string)] = i
	}
	for _, i := range cl.Included {
		if i["@type"].(string) != "Component" {
			continue
		}
		cr := componentResponse{}
		cr.Name = i["name"].(string)
		cr.Namespace = i["namespace"].(string)
		cr.ComponentType = i["component_type"].(string)
		cr.ID = fmt.Sprintf("%s.%s", cr.Name, cr.Namespace)
		cr.Description = i["description"].(string)
		cr.Source = i["source"].(string)
		cr.Clusters = make([]string, 0)
		for _, c := range i["clusters"].([]interface{}) {
			cr.Clusters = append(cr.Clusters, c.(map[string]interface{})["cluster_id"].(string))
		}
		cr.ResilienceZones = make([]resilienceZoneResponse, 0)
		for _, rz := range i["resilience_zones"].([]interface{}) {
			cr.ResilienceZones = append(cr.ResilienceZones, resilienceZoneResponse{
				Name:    rz.(map[string]interface{})["name"].(string),
				Version: rz.(map[string]interface{})["version"].(string),
			})
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
	Logger logger.Logger
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
		return nil, fmt.Errorf("apiclient: %w", err)
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

func listComponents(ctx context.Context, in *ListComponentsInput, componentList *ComponentList) (*apiclient.Problem, error) {
	response, err := in.APIClient.ListComponentsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading components: %w", err)
	}
	in.Logger.Debug("apiclient",
		"status", response.StatusCode(),
		"content-type", response.HTTPResponse.Header.Get("Content-Type"))
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
	Logger    logger.Logger
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
		return nil, fmt.Errorf("apiclient: %w", err)
	}
	in.Logger.Debug("apiclient",
		"status", response.StatusCode(),
		"content-type", response.HTTPResponse.Header.Get("Content-Type"))
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
	includeMap := make(map[string]interface{})
	if component.Included != nil {
		for _, i := range *component.Included {
			includeMap[i["@id"].(string)] = i
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
	for _, rz := range *component.ResilienceZones {
		cr.ResilienceZones = append(cr.ResilienceZones, resilienceZoneResponse{
			Name:    nilStr(rz.Name),
			Version: nilStr(rz.Version),
		})
	}
	return cr
}

func nilStr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
