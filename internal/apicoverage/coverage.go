package apicoverage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"

	jules "github.com/SamyRai/go-jules"
)

const DefaultDiscoveryURL = "https://jules.googleapis.com/$discovery/rest?version=v1alpha"

// Report summarizes discovery coverage validation.
type Report struct {
	Revision       string
	OperationCount int
	SchemaCount    int
	EnumCount      int
}

type discoveryDoc struct {
	Revision  string                       `json:"revision"`
	Resources map[string]discoveryResource `json:"resources"`
	Schemas   map[string]discoverySchema   `json:"schemas"`
}

type discoveryResource struct {
	Methods   map[string]discoveryMethod   `json:"methods"`
	Resources map[string]discoveryResource `json:"resources"`
}

type discoveryMethod struct {
	HTTPMethod string                        `json:"httpMethod"`
	Path       string                        `json:"path"`
	Parameters map[string]discoveryParameter `json:"parameters"`
}

type discoveryParameter struct {
	Location string `json:"location"`
}

type discoverySchema struct {
	Properties map[string]discoveryProperty `json:"properties"`
}

type discoveryProperty struct {
	Enum []string `json:"enum"`
}

type operationCoverage struct {
	Resource  string
	Method    string
	Service   string
	SDKMethod string
	HTTP      string
	Path      string
	Params    []string
}

// ValidateURL fetches a Jules Discovery document and validates the SDK coverage
// table against it.
func ValidateURL(ctx context.Context, discoveryURL string) (*Report, error) {
	if discoveryURL == "" {
		discoveryURL = DefaultDiscoveryURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create discovery request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch discovery document: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch discovery document: HTTP %d", resp.StatusCode)
	}
	return ValidateReader(resp.Body)
}

// ValidateReader validates a Discovery document read from r.
func ValidateReader(r io.Reader) (*Report, error) {
	var doc discoveryDoc
	if err := json.NewDecoder(r).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode discovery document: %w", err)
	}
	var errs []string

	operationCount := validateOperations(doc, &errs)
	schemaCount := validateSchemas(doc, &errs)
	enumCount := validateEnums(doc, &errs)

	if len(errs) > 0 {
		sort.Strings(errs)
		return nil, fmt.Errorf("Jules API coverage failed:\n- %s", strings.Join(errs, "\n- "))
	}

	return &Report{
		Revision:       doc.Revision,
		OperationCount: operationCount,
		SchemaCount:    schemaCount,
		EnumCount:      enumCount,
	}, nil
}

func validateOperations(doc discoveryDoc, errs *[]string) int {
	coverage := []operationCoverage{
		{Resource: "sessions", Method: "get", Service: "sessions", SDKMethod: "Get", HTTP: "GET", Path: "v1alpha/{+name}", Params: []string{"name"}},
		{Resource: "sessions", Method: "list", Service: "sessions", SDKMethod: "List", HTTP: "GET", Path: "v1alpha/sessions", Params: []string{"filter", "pageSize", "pageToken"}},
		{Resource: "sessions", Method: "create", Service: "sessions", SDKMethod: "Create", HTTP: "POST", Path: "v1alpha/sessions"},
		{Resource: "sessions", Method: "delete", Service: "sessions", SDKMethod: "Delete", HTTP: "DELETE", Path: "v1alpha/{+name}", Params: []string{"name"}},
		{Resource: "sessions", Method: "sendMessage", Service: "sessions", SDKMethod: "SendMessage", HTTP: "POST", Path: "v1alpha/{+session}:sendMessage", Params: []string{"session"}},
		{Resource: "sessions", Method: "approvePlan", Service: "sessions", SDKMethod: "ApprovePlan", HTTP: "POST", Path: "v1alpha/{+session}:approvePlan", Params: []string{"session"}},
		{Resource: "sessions", Method: "archive", Service: "sessions", SDKMethod: "Archive", HTTP: "POST", Path: "v1alpha/{+name}:archive", Params: []string{"name"}},
		{Resource: "sessions", Method: "unarchive", Service: "sessions", SDKMethod: "Unarchive", HTTP: "POST", Path: "v1alpha/{+name}:unarchive", Params: []string{"name"}},
		{Resource: "sessions.activities", Method: "get", Service: "activities", SDKMethod: "Get", HTTP: "GET", Path: "v1alpha/{+name}", Params: []string{"name"}},
		{Resource: "sessions.activities", Method: "list", Service: "activities", SDKMethod: "List", HTTP: "GET", Path: "v1alpha/{+parent}/activities", Params: []string{"filter", "pageSize", "pageToken", "parent"}},
		{Resource: "sources", Method: "get", Service: "sources", SDKMethod: "Get", HTTP: "GET", Path: "v1alpha/{+name}", Params: []string{"name"}},
		{Resource: "sources", Method: "list", Service: "sources", SDKMethod: "List", HTTP: "GET", Path: "v1alpha/sources", Params: []string{"filter", "pageSize", "pageToken"}},
	}

	expected := make(map[string]operationCoverage, len(coverage))
	for _, op := range coverage {
		expected[op.Resource+"."+op.Method] = op
	}

	serviceTypes := map[string]reflect.Type{
		"activities": reflect.TypeOf((*jules.ActivitiesService)(nil)),
		"sessions":   reflect.TypeOf((*jules.SessionsService)(nil)),
		"sources":    reflect.TypeOf((*jules.SourcesService)(nil)),
	}
	seen := make(map[string]struct{})
	for _, resourceName := range sortedResourceNames(doc.Resources) {
		resource, ok := resourceByName(doc.Resources, resourceName)
		if !ok {
			*errs = append(*errs, "missing resource "+resourceName)
			continue
		}
		for _, methodName := range sortedMethodNames(resource.Methods) {
			key := resourceName + "." + methodName
			seen[key] = struct{}{}
			op, ok := expected[key]
			if !ok {
				*errs = append(*errs, "uncovered operation "+key)
				continue
			}
			method := resource.Methods[methodName]
			if method.HTTPMethod != op.HTTP {
				*errs = append(*errs, fmt.Sprintf("%s HTTP method = %q, want %q", key, method.HTTPMethod, op.HTTP))
			}
			if method.Path != op.Path {
				*errs = append(*errs, fmt.Sprintf("%s path = %q, want %q", key, method.Path, op.Path))
			}
			if missing := missingKeys(method.Parameters, op.Params); len(missing) > 0 {
				*errs = append(*errs, fmt.Sprintf("%s missing documented params: %s", key, strings.Join(missing, ", ")))
			}
			if extra := extraKeys(method.Parameters, op.Params); len(extra) > 0 {
				*errs = append(*errs, fmt.Sprintf("%s has uncovered params: %s", key, strings.Join(extra, ", ")))
			}
			serviceType, ok := serviceTypes[op.Service]
			if !ok {
				*errs = append(*errs, fmt.Sprintf("%s unknown SDK service %s", key, op.Service))
				continue
			}
			if _, ok := serviceType.MethodByName(op.SDKMethod); !ok {
				*errs = append(*errs, fmt.Sprintf("%s missing SDK method %s.%s", key, op.Service, op.SDKMethod))
			}
		}
	}

	for key := range expected {
		if _, ok := seen[key]; !ok {
			*errs = append(*errs, "expected operation missing from discovery: "+key)
		}
	}

	return len(coverage)
}

func validateSchemas(doc discoveryDoc, errs *[]string) int {
	schemaTypes := map[string]reflect.Type{
		"Session":                reflect.TypeOf(jules.Session{}),
		"SourceContext":          reflect.TypeOf(jules.SourceContext{}),
		"GitHubRepoContext":      reflect.TypeOf(jules.GithubRepoContext{}),
		"SessionOutput":          reflect.TypeOf(jules.Output{}),
		"PullRequest":            reflect.TypeOf(jules.PullRequest{}),
		"ChangeSet":              reflect.TypeOf(jules.ChangeSet{}),
		"GitPatch":               reflect.TypeOf(jules.GitPatch{}),
		"ListSessionsResponse":   reflect.TypeOf(jules.SessionsResponse{}),
		"Activity":               reflect.TypeOf(jules.Activity{}),
		"AgentMessaged":          reflect.TypeOf(jules.AgentMessaged{}),
		"UserMessaged":           reflect.TypeOf(jules.UserMessaged{}),
		"PlanGenerated":          reflect.TypeOf(jules.PlanGenerated{}),
		"Plan":                   reflect.TypeOf(jules.Plan{}),
		"PlanStep":               reflect.TypeOf(jules.Step{}),
		"PlanApproved":           reflect.TypeOf(jules.PlanApproved{}),
		"ProgressUpdated":        reflect.TypeOf(jules.ProgressUpdated{}),
		"SessionCompleted":       reflect.TypeOf(jules.SessionCompleted{}),
		"SessionFailed":          reflect.TypeOf(jules.SessionFailed{}),
		"Artifact":               reflect.TypeOf(jules.Artifact{}),
		"Media":                  reflect.TypeOf(jules.Media{}),
		"BashOutput":             reflect.TypeOf(jules.BashOutput{}),
		"ListActivitiesResponse": reflect.TypeOf(jules.ActivitiesResponse{}),
		"SendMessageRequest":     reflect.TypeOf(jules.SendMessageRequest{}),
		"Source":                 reflect.TypeOf(jules.Source{}),
		"GitHubRepo":             reflect.TypeOf(jules.GithubRepo{}),
		"GitHubBranch":           reflect.TypeOf(jules.Branch{}),
		"ListSourcesResponse":    reflect.TypeOf(jules.SourcesResponse{}),
	}
	operationOnlySchemas := map[string]struct{}{
		"ApprovePlanRequest":      {},
		"ApprovePlanResponse":     {},
		"ArchiveSessionRequest":   {},
		"Empty":                   {},
		"SendMessageResponse":     {},
		"UnarchiveSessionRequest": {},
	}

	covered := 0
	for schemaName, schema := range doc.Schemas {
		if _, ok := operationOnlySchemas[schemaName]; ok {
			covered++
			continue
		}
		goType, ok := schemaTypes[schemaName]
		if !ok {
			*errs = append(*errs, "unmapped schema "+schemaName)
			continue
		}
		jsonFields := jsonFieldNames(goType)
		for property := range schema.Properties {
			if _, ok := jsonFields[property]; !ok {
				*errs = append(*errs, fmt.Sprintf("%s missing JSON field %q", schemaName, property))
			}
		}
		covered++
	}

	return covered
}

func validateEnums(doc discoveryDoc, errs *[]string) int {
	sessionStateSchema := doc.Schemas["Session"].Properties["state"]
	assertEnumValues("Session.state", sessionStateSchema.Enum, []string{
		string(jules.SessionStateUnspecified),
		string(jules.SessionStateQueued),
		string(jules.SessionStatePlanning),
		string(jules.SessionStateAwaitingPlanApproval),
		string(jules.SessionStateAwaitingUserFeedback),
		string(jules.SessionStateInProgress),
		string(jules.SessionStatePaused),
		string(jules.SessionStateFailed),
		string(jules.SessionStateCompleted),
	}, errs)

	automationModeSchema := doc.Schemas["Session"].Properties["automationMode"]
	assertEnumValues("Session.automationMode", automationModeSchema.Enum, []string{
		string(jules.AutomationModeUnspecified),
		string(jules.AutomationModeAutoCreatePR),
	}, errs)

	return 2
}

func assertEnumValues(name string, documented, implemented []string, errs *[]string) {
	implementedSet := make(map[string]struct{}, len(implemented))
	for _, value := range implemented {
		implementedSet[value] = struct{}{}
	}
	for _, value := range documented {
		if _, ok := implementedSet[value]; !ok {
			*errs = append(*errs, fmt.Sprintf("%s missing enum value %q", name, value))
		}
	}
}

func jsonFieldNames(t reflect.Type) map[string]struct{} {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	fields := make(map[string]struct{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" {
			name = field.Name
		}
		if name != "-" {
			fields[name] = struct{}{}
		}
	}
	return fields
}

func sortedResourceNames(resources map[string]discoveryResource) []string {
	var names []string
	for name, resource := range resources {
		names = append(names, name)
		for nested := range resource.Resources {
			names = append(names, name+"."+nested)
		}
	}
	sort.Strings(names)
	return names
}

func resourceByName(resources map[string]discoveryResource, name string) (discoveryResource, bool) {
	parts := strings.Split(name, ".")
	resource, ok := resources[parts[0]]
	if !ok {
		return discoveryResource{}, false
	}
	for _, part := range parts[1:] {
		resource, ok = resource.Resources[part]
		if !ok {
			return discoveryResource{}, false
		}
	}
	return resource, true
}

func sortedMethodNames(methods map[string]discoveryMethod) []string {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func missingKeys(parameters map[string]discoveryParameter, expected []string) []string {
	var missing []string
	for _, key := range expected {
		if _, ok := parameters[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}

func extraKeys(parameters map[string]discoveryParameter, expected []string) []string {
	expectedSet := make(map[string]struct{}, len(expected))
	for _, key := range expected {
		expectedSet[key] = struct{}{}
	}
	var extra []string
	for key := range parameters {
		if _, ok := expectedSet[key]; !ok {
			extra = append(extra, key)
		}
	}
	sort.Strings(extra)
	return extra
}
