package apicoverage

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateReaderAcceptsCoveredDiscovery(t *testing.T) {
	report, err := ValidateReader(mustDiscoveryJSON(t, testDiscoveryDoc()))
	if err != nil {
		t.Fatalf("ValidateReader returned error: %v", err)
	}
	if report.Revision != "test-revision" {
		t.Fatalf("Revision = %q, want test-revision", report.Revision)
	}
	if report.OperationCount != 12 {
		t.Fatalf("OperationCount = %d, want 12", report.OperationCount)
	}
	if report.EnumCount != 2 {
		t.Fatalf("EnumCount = %d, want 2", report.EnumCount)
	}
}

func TestValidateReaderRejectsUncoveredOperation(t *testing.T) {
	doc := testDiscoveryDoc()
	doc.Resources["sessions"].Methods["pause"] = discoveryMethod{
		HTTPMethod: "POST",
		Path:       "v1alpha/{+name}:pause",
		Parameters: map[string]discoveryParameter{"name": {Location: "path"}},
	}

	_, err := ValidateReader(mustDiscoveryJSON(t, doc))
	if err == nil {
		t.Fatal("ValidateReader returned nil error")
	}
	if !strings.Contains(err.Error(), "uncovered operation sessions.pause") {
		t.Fatalf("error = %q, want uncovered operation", err)
	}
}

func TestValidateReaderRejectsMissingSchemaField(t *testing.T) {
	doc := testDiscoveryDoc()
	doc.Schemas["Session"].Properties["newField"] = discoveryProperty{}

	_, err := ValidateReader(mustDiscoveryJSON(t, doc))
	if err == nil {
		t.Fatal("ValidateReader returned nil error")
	}
	if !strings.Contains(err.Error(), `Session missing JSON field "newField"`) {
		t.Fatalf("error = %q, want missing schema field", err)
	}
}

func mustDiscoveryJSON(t *testing.T, doc discoveryDoc) *bytes.Reader {
	t.Helper()
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal discovery fixture: %v", err)
	}
	return bytes.NewReader(data)
}

func testDiscoveryDoc() discoveryDoc {
	return discoveryDoc{
		Revision: "test-revision",
		Resources: map[string]discoveryResource{
			"sessions": {
				Methods: map[string]discoveryMethod{
					"get":         method("GET", "v1alpha/{+name}", "name"),
					"list":        method("GET", "v1alpha/sessions", "filter", "pageSize", "pageToken"),
					"create":      method("POST", "v1alpha/sessions"),
					"delete":      method("DELETE", "v1alpha/{+name}", "name"),
					"sendMessage": method("POST", "v1alpha/{+session}:sendMessage", "session"),
					"approvePlan": method("POST", "v1alpha/{+session}:approvePlan", "session"),
					"archive":     method("POST", "v1alpha/{+name}:archive", "name"),
					"unarchive":   method("POST", "v1alpha/{+name}:unarchive", "name"),
				},
				Resources: map[string]discoveryResource{
					"activities": {
						Methods: map[string]discoveryMethod{
							"get":  method("GET", "v1alpha/{+name}", "name"),
							"list": method("GET", "v1alpha/{+parent}/activities", "filter", "pageSize", "pageToken", "parent"),
						},
					},
				},
			},
			"sources": {
				Methods: map[string]discoveryMethod{
					"get":  method("GET", "v1alpha/{+name}", "name"),
					"list": method("GET", "v1alpha/sources", "filter", "pageSize", "pageToken"),
				},
			},
		},
		Schemas: map[string]discoverySchema{
			"Session": schemaWithEnums(
				map[string][]string{
					"state": {
						"STATE_UNSPECIFIED",
						"QUEUED",
						"PLANNING",
						"AWAITING_PLAN_APPROVAL",
						"AWAITING_USER_FEEDBACK",
						"IN_PROGRESS",
						"PAUSED",
						"FAILED",
						"COMPLETED",
					},
					"automationMode": {
						"AUTOMATION_MODE_UNSPECIFIED",
						"AUTO_CREATE_PR",
					},
				},
				"name", "id", "prompt", "sourceContext", "title", "requirePlanApproval", "automationMode", "createTime", "updateTime", "state", "url", "outputs", "archived",
			),
			"SourceContext":           schema("githubRepoContext", "source", "workingBranch", "environmentVariablesEnabled"),
			"GitHubRepoContext":       schema("startingBranch"),
			"SessionOutput":           schema("pullRequest", "changeSet"),
			"PullRequest":             schema("url", "title", "description", "baseRef", "headRef"),
			"ChangeSet":               schema("source", "gitPatch"),
			"GitPatch":                schema("unidiffPatch", "baseCommitId", "suggestedCommitMessage"),
			"ListSessionsResponse":    schema("sessions", "nextPageToken"),
			"Activity":                schema("name", "id", "description", "createTime", "originator", "artifacts", "agentMessaged", "userMessaged", "planGenerated", "planApproved", "progressUpdated", "sessionCompleted", "sessionFailed"),
			"AgentMessaged":           schema("agentMessage"),
			"UserMessaged":            schema("userMessage"),
			"PlanGenerated":           schema("plan"),
			"Plan":                    schema("id", "steps", "createTime"),
			"PlanStep":                schema("id", "title", "description", "index"),
			"PlanApproved":            schema("planId"),
			"ProgressUpdated":         schema("title", "description"),
			"SessionCompleted":        schema(),
			"SessionFailed":           schema("reason"),
			"Artifact":                schema("changeSet", "media", "bashOutput"),
			"Media":                   schema("data", "mimeType"),
			"BashOutput":              schema("command", "output", "exitCode"),
			"ListActivitiesResponse":  schema("activities", "nextPageToken"),
			"SendMessageRequest":      schema("prompt"),
			"SendMessageResponse":     schema(),
			"ApprovePlanRequest":      schema(),
			"ApprovePlanResponse":     schema(),
			"ArchiveSessionRequest":   schema(),
			"UnarchiveSessionRequest": schema(),
			"Empty":                   schema(),
			"Source":                  schema("name", "id", "githubRepo"),
			"GitHubRepo":              schema("owner", "repo", "isPrivate", "defaultBranch", "branches"),
			"GitHubBranch":            schema("displayName"),
			"ListSourcesResponse":     schema("sources", "nextPageToken"),
		},
	}
}

func method(httpMethod, path string, params ...string) discoveryMethod {
	parameters := make(map[string]discoveryParameter, len(params))
	for _, param := range params {
		parameters[param] = discoveryParameter{Location: "query"}
	}
	return discoveryMethod{
		HTTPMethod: httpMethod,
		Path:       path,
		Parameters: parameters,
	}
}

func schema(properties ...string) discoverySchema {
	return schemaWithEnums(nil, properties...)
}

func schemaWithEnums(enums map[string][]string, properties ...string) discoverySchema {
	schema := discoverySchema{Properties: make(map[string]discoveryProperty, len(properties))}
	for _, property := range properties {
		schema.Properties[property] = discoveryProperty{Enum: enums[property]}
	}
	return schema
}
