package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/SamyRai/go-jules/internal/model"
	"github.com/SamyRai/go-jules/internal/resource"
)

// ListActivitiesOptions controls pagination and official API filters.
type ListActivitiesOptions struct {
	PageSize   int
	PageToken  string
	Filter     string
	CreateTime time.Time
}

// ActivityFilter represents client-side filters for documented activity data.
type ActivityFilter struct {
	Type         string
	Status       string
	CreateTime   time.Time
	Before       time.Time
	After        time.Time
	HasPlan      *bool
	HasArtifacts *bool
}

// ActivitySearchOptions represents client-side search options for activities.
type ActivitySearchOptions struct {
	Query  string
	Filter *ActivityFilter
	Limit  int
}

// ActivitiesService owns Jules activity operations.
type ActivitiesService struct {
	baseURL string
	doer    jsonDoer
}

func NewActivitiesService(baseURL string, doer jsonDoer) *ActivitiesService {
	return &ActivitiesService{baseURL: baseURL, doer: doer}
}

// List lists activities with pagination. When CreateTime is set, the request
// includes the documented createTime cursor filter and the response is filtered
// client-side as a defensive guard.
func (a *ActivitiesService) List(ctx context.Context, sessionID string, options *ListActivitiesOptions) (*model.ActivitiesResponse, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	pageSize := 50
	pageToken := ""
	filter := ""
	createTime := time.Time{}
	if options != nil {
		pageSize = options.PageSize
		pageToken = options.PageToken
		filter = options.Filter
		createTime = options.CreateTime
	}
	pageSize = normalizePageSize(pageSize, 50, 100)

	resourcePath, err := resource.SessionPath(sessionID)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}
	if filter != "" {
		query.Set("filter", filter)
	}
	if !createTime.IsZero() {
		query.Set("createTime", createTime.Format(time.RFC3339Nano))
	}

	requestURL := fmt.Sprintf("%s/%s/activities?%s", a.baseURL, resourcePath, query.Encode())

	var response model.ActivitiesResponse
	if err := a.doer.DoJSON(ctx, "GET", requestURL, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}
	if !createTime.IsZero() {
		response.Activities = activitiesAtOrAfter(response.Activities, createTime)
	}

	return &response, nil
}

// ListAll retrieves every activity by following nextPageToken.
func (a *ActivitiesService) ListAll(ctx context.Context, sessionID string, pageSize int) ([]model.Activity, error) {
	var activities []model.Activity
	pageToken := ""
	for {
		response, err := a.List(ctx, sessionID, &ListActivitiesOptions{
			PageSize:  pageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}
		activities = append(activities, response.Activities...)
		if response.NextPageToken == "" {
			return activities, nil
		}
		pageToken = response.NextPageToken
	}
}

// ListSince retrieves activities created at or after the cursor time.
func (a *ActivitiesService) ListSince(ctx context.Context, sessionID string, cursor time.Time, pageSize int) ([]model.Activity, error) {
	var activities []model.Activity
	pageToken := ""
	for {
		response, err := a.List(ctx, sessionID, &ListActivitiesOptions{
			PageSize:   pageSize,
			PageToken:  pageToken,
			CreateTime: cursor,
		})
		if err != nil {
			return nil, err
		}
		activities = append(activities, response.Activities...)
		if response.NextPageToken == "" {
			return activities, nil
		}
		pageToken = response.NextPageToken
	}
}

func activitiesAtOrAfter(activities []model.Activity, cursor time.Time) []model.Activity {
	if cursor.IsZero() {
		return activities
	}
	filtered := make([]model.Activity, 0, len(activities))
	for _, activity := range activities {
		if activity.CreateTime.IsZero() {
			continue
		}
		if !activity.CreateTime.Before(cursor) {
			filtered = append(filtered, activity)
		}
	}
	return filtered
}

// ActivityCursor returns the latest createTime in the provided activities.
func ActivityCursor(activities []model.Activity) time.Time {
	var cursor time.Time
	for _, activity := range activities {
		if activity.CreateTime.After(cursor) {
			cursor = activity.CreateTime
		}
	}
	return cursor
}

// Get retrieves a specific activity by ID or resource name.
func (a *ActivitiesService) Get(ctx context.Context, sessionID, activityID string) (*model.Activity, error) {
	if activityID == "" {
		return nil, fmt.Errorf("activity ID is required")
	}

	resourcePath, err := resource.ActivityPath(sessionID, activityID)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s/%s", a.baseURL, resourcePath)

	var activity model.Activity
	if err := a.doer.DoJSON(ctx, "GET", requestURL, nil, &activity); err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
}

// Filter lists activities and applies client-side filters over documented fields.
func (a *ActivitiesService) Filter(ctx context.Context, sessionID string, filter *ActivityFilter) ([]model.Activity, error) {
	options := &ListActivitiesOptions{}
	if filter != nil {
		options.CreateTime = filter.CreateTime
		if options.CreateTime.IsZero() {
			options.CreateTime = filter.After
		}
	}

	response, err := a.List(ctx, sessionID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list filtered activities: %w", err)
	}

	return FilterActivities(response.Activities, filter), nil
}

// Search searches documented activity payloads client-side. It does
// not call an undocumented search endpoint.
func (a *ActivitiesService) Search(ctx context.Context, sessionID string, options *ActivitySearchOptions) ([]model.Activity, error) {
	activities, err := a.ListAll(ctx, sessionID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities for search: %w", err)
	}

	if options == nil {
		return activities, nil
	}
	activities = FilterActivities(activities, options.Filter)
	if options.Query != "" {
		activities = searchActivityPayloads(activities, options.Query)
	}
	if options.Limit > 0 && len(activities) > options.Limit {
		activities = activities[:options.Limit]
	}
	return activities, nil
}

// FilterActivities applies client-side filters over documented activity fields.
func FilterActivities(activities []model.Activity, filter *ActivityFilter) []model.Activity {
	if filter == nil {
		return activities
	}

	filtered := make([]model.Activity, 0, len(activities))
	for _, activity := range activities {
		if filter.Type != "" && !activityMatchesType(activity, filter.Type) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(activity.Status, filter.Status) {
			continue
		}
		if !filter.Before.IsZero() && !activity.CreateTime.IsZero() && !activity.CreateTime.Before(filter.Before) {
			continue
		}
		if !filter.After.IsZero() && !activity.CreateTime.IsZero() && activity.CreateTime.Before(filter.After) {
			continue
		}
		if filter.HasPlan != nil && activityHasPlan(activity) != *filter.HasPlan {
			continue
		}
		if filter.HasArtifacts != nil && (len(activity.Artifacts) > 0) != *filter.HasArtifacts {
			continue
		}
		filtered = append(filtered, activity)
	}

	return filtered
}

func activityMatchesType(activity model.Activity, activityType string) bool {
	normalized := strings.ToLower(activityType)
	switch {
	case strings.Contains(normalized, "plan") && activityHasPlan(activity):
		return true
	case strings.Contains(normalized, "message") && (activity.UserMessaged != nil || activity.AgentMessaged != nil):
		return true
	case strings.Contains(normalized, "user") && activity.UserMessaged != nil:
		return true
	case strings.Contains(normalized, "agent") && activity.AgentMessaged != nil:
		return true
	case strings.Contains(normalized, "progress") && activity.ProgressUpdated != nil:
		return true
	case strings.Contains(normalized, "complete") && activity.SessionCompleted != nil:
		return true
	case strings.Contains(normalized, "fail") && activity.SessionFailed != nil:
		return true
	case strings.Contains(normalized, "artifact") && len(activity.Artifacts) > 0:
		return true
	}

	return strings.Contains(strings.ToLower(activitySearchText(activity)), normalized)
}

func activityHasPlan(activity model.Activity) bool {
	return activity.PlanGenerated != nil || activity.PlanApproved != nil
}

func searchActivityPayloads(activities []model.Activity, query string) []model.Activity {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return activities
	}
	filtered := make([]model.Activity, 0, len(activities))
	for _, activity := range activities {
		if strings.Contains(strings.ToLower(activitySearchText(activity)), query) {
			filtered = append(filtered, activity)
		}
	}
	return filtered
}

func activitySearchText(activity model.Activity) string {
	parts := []string{
		activity.Name,
		activity.Description,
		string(activity.Originator),
		activity.Status,
	}
	if activity.UserMessaged != nil {
		parts = append(parts, activity.UserMessaged.UserMessage)
	}
	if activity.AgentMessaged != nil {
		parts = append(parts, activity.AgentMessaged.AgentMessage)
	}
	if activity.ProgressUpdated != nil {
		parts = append(parts, activity.ProgressUpdated.Title, activity.ProgressUpdated.Description)
	}
	if activity.SessionFailed != nil {
		parts = append(parts, activity.SessionFailed.Reason)
	}
	for _, artifact := range activity.Artifacts {
		if artifact.BashOutput != nil {
			parts = append(parts, artifact.BashOutput.Command, artifact.BashOutput.Output)
		}
		if artifact.ChangeSet != nil {
			parts = append(parts, artifact.ChangeSet.Source)
			if artifact.ChangeSet.GitPatch != nil {
				parts = append(parts, artifact.ChangeSet.GitPatch.SuggestedCommitMessage, artifact.ChangeSet.GitPatch.UnidiffPatch)
			}
		}
		if artifact.Media != nil {
			parts = append(parts, artifact.Media.MimeType)
		}
	}
	return strings.Join(parts, " ")
}
