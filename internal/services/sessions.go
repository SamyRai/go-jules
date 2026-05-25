package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/SamyRai/go-jules/internal/model"
	"github.com/SamyRai/go-jules/internal/resource"
)

// ListSessionsOptions controls session pagination and filtering.
type ListSessionsOptions struct {
	PageSize  int
	PageToken string
	Filter    string
}

// SessionsService owns Jules session operations.
type SessionsService struct {
	baseURL string
	doer    jsonDoer
	sources SourceResolver
}

// SourceResolver is the narrow dependency needed to infer source defaults.
type SourceResolver interface {
	Get(ctx context.Context, sourceID string) (*model.Source, error)
}

func NewSessionsService(baseURL string, doer jsonDoer, sources SourceResolver) *SessionsService {
	return &SessionsService{baseURL: baseURL, doer: doer, sources: sources}
}

// Create creates a new coding session.
func (s *SessionsService) Create(ctx context.Context, req *model.CreateSessionRequest) (*model.Session, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	req = cloneCreateSessionRequest(req)
	if req.SourceContext != nil && req.SourceContext.Source != "" {
		req.SourceContext.Source = resource.NormalizeSourceName(req.SourceContext.Source)
		if req.SourceContext.GithubRepoContext == nil || req.SourceContext.GithubRepoContext.StartingBranch == "" {
			branch, err := s.defaultStartingBranch(ctx, req.SourceContext.Source)
			if err != nil {
				return nil, err
			}
			req.SourceContext.GithubRepoContext = &model.GithubRepoContext{StartingBranch: branch}
		}
	}

	requestURL := fmt.Sprintf("%s/sessions", s.baseURL)

	var session model.Session
	if err := s.doer.DoJSON(ctx, "POST", requestURL, req, &session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	if session.ID != "" && session.State == "" {
		if hydrated, err := s.Get(ctx, session.ID); err == nil {
			return hydrated, nil
		}
	}

	return &session, nil
}

// Get retrieves a specific session by ID.
func (s *SessionsService) Get(ctx context.Context, sessionID string) (*model.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	resourcePath, err := resource.SessionPath(sessionID)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s/%s", s.baseURL, resourcePath)

	var session model.Session
	if err := s.doer.DoJSON(ctx, "GET", requestURL, nil, &session); err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// Delete deletes a session by ID.
func (s *SessionsService) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	resourcePath, err := resource.SessionPath(sessionID)
	if err != nil {
		return err
	}
	requestURL := fmt.Sprintf("%s/%s", s.baseURL, resourcePath)
	if err := s.doer.DoJSON(ctx, "DELETE", requestURL, nil, nil); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// List lists sessions with pagination and official API filters.
func (s *SessionsService) List(ctx context.Context, options *ListSessionsOptions) (*model.SessionsResponse, error) {
	pageSize := 30
	pageToken := ""
	filter := ""
	if options != nil {
		pageSize = options.PageSize
		pageToken = options.PageToken
		filter = options.Filter
	}
	pageSize = normalizePageSize(pageSize, 30, 100)

	query := url.Values{}
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}
	if filter != "" {
		query.Set("filter", filter)
	}
	requestURL := fmt.Sprintf("%s/sessions?%s", s.baseURL, query.Encode())

	var response model.SessionsResponse
	if err := s.doer.DoJSON(ctx, "GET", requestURL, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return &response, nil
}

// ListAll retrieves every session matching the official filter expression by
// following nextPageToken.
func (s *SessionsService) ListAll(ctx context.Context, pageSize int, filter string) ([]model.Session, error) {
	var sessions []model.Session
	pageToken := ""
	for {
		response, err := s.List(ctx, &ListSessionsOptions{
			PageSize:  pageSize,
			PageToken: pageToken,
			Filter:    filter,
		})
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, response.Sessions...)
		if response.NextPageToken == "" {
			return sessions, nil
		}
		pageToken = response.NextPageToken
	}
}

// Archive archives a session by ID and returns the updated session.
func (s *SessionsService) Archive(ctx context.Context, sessionID string) (*model.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	resourcePath, err := resource.SessionPath(sessionID)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s/%s:archive", s.baseURL, resourcePath)

	var session model.Session
	if err := s.doer.DoJSON(ctx, "POST", requestURL, map[string]any{}, &session); err != nil {
		return nil, fmt.Errorf("failed to archive session: %w", err)
	}

	return &session, nil
}

// Unarchive unarchives a session by ID and returns the updated session.
func (s *SessionsService) Unarchive(ctx context.Context, sessionID string) (*model.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	resourcePath, err := resource.SessionPath(sessionID)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s/%s:unarchive", s.baseURL, resourcePath)

	var session model.Session
	if err := s.doer.DoJSON(ctx, "POST", requestURL, map[string]any{}, &session); err != nil {
		return nil, fmt.Errorf("failed to unarchive session: %w", err)
	}

	return &session, nil
}

// SendMessage sends a message to Jules within a session.
func (s *SessionsService) SendMessage(ctx context.Context, sessionID string, req *model.SendMessageRequest) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if req == nil || req.Prompt == "" {
		return fmt.Errorf("message prompt is required")
	}

	resourcePath, err := resource.SessionPath(sessionID)
	if err != nil {
		return err
	}
	requestURL := fmt.Sprintf("%s/%s:sendMessage", s.baseURL, resourcePath)

	if err := s.doer.DoJSON(ctx, "POST", requestURL, req, nil); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// ApprovePlan approves a plan in a session.
func (s *SessionsService) ApprovePlan(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	resourcePath, err := resource.SessionPath(sessionID)
	if err != nil {
		return err
	}
	requestURL := fmt.Sprintf("%s/%s:approvePlan", s.baseURL, resourcePath)

	if err := s.doer.DoJSON(ctx, "POST", requestURL, nil, nil); err != nil {
		return fmt.Errorf("failed to approve plan: %w", err)
	}

	return nil
}

func (s *SessionsService) defaultStartingBranch(ctx context.Context, sourceName string) (string, error) {
	source, err := s.sources.Get(ctx, sourceName)
	if err != nil {
		return "", fmt.Errorf("failed to infer starting branch for %s: %w", sourceName, err)
	}
	if source.GithubRepo != nil {
		if source.GithubRepo.DefaultBranch != nil && source.GithubRepo.DefaultBranch.DisplayName != "" {
			return source.GithubRepo.DefaultBranch.DisplayName, nil
		}
		for _, branch := range source.GithubRepo.Branches {
			if branch.DisplayName != "" {
				return branch.DisplayName, nil
			}
		}
	}
	return "", fmt.Errorf("starting branch is required for source-backed sessions; pass githubRepoContext.startingBranch")
}

func cloneCreateSessionRequest(req *model.CreateSessionRequest) *model.CreateSessionRequest {
	clone := *req
	if req.SourceContext != nil {
		sourceContext := *req.SourceContext
		if req.SourceContext.GithubRepoContext != nil {
			githubRepoContext := *req.SourceContext.GithubRepoContext
			sourceContext.GithubRepoContext = &githubRepoContext
		}
		clone.SourceContext = &sourceContext
	}
	return &clone
}
