package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/SamyRai/go-jules/internal/model"
	"github.com/SamyRai/go-jules/internal/resource"
)

type jsonDoer interface {
	DoJSON(ctx context.Context, method, url string, body any, result any) error
}

// ListSourcesOptions controls source pagination and filtering.
type ListSourcesOptions struct {
	PageSize  int
	PageToken string
	Filter    string
}

// SourcesService owns Jules source operations.
type SourcesService struct {
	baseURL string
	doer    jsonDoer
}

func NewSourcesService(baseURL string, doer jsonDoer) *SourcesService {
	return &SourcesService{baseURL: baseURL, doer: doer}
}

// List lists available code sources with pagination and filtering.
func (s *SourcesService) List(ctx context.Context, options *ListSourcesOptions) (*model.SourcesResponse, error) {
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
	requestURL := fmt.Sprintf("%s/sources?%s", s.baseURL, query.Encode())

	var response model.SourcesResponse
	if err := s.doer.DoJSON(ctx, "GET", requestURL, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}

	return &response, nil
}

// ListAll retrieves every source by following nextPageToken.
func (s *SourcesService) ListAll(ctx context.Context, pageSize int, filter string) ([]model.Source, error) {
	var sources []model.Source
	pageToken := ""
	for {
		response, err := s.List(ctx, &ListSourcesOptions{
			PageSize:  pageSize,
			PageToken: pageToken,
			Filter:    filter,
		})
		if err != nil {
			return nil, err
		}
		sources = append(sources, response.Sources...)
		if response.NextPageToken == "" {
			return sources, nil
		}
		pageToken = response.NextPageToken
	}
}

// Get retrieves a specific source by ID.
func (s *SourcesService) Get(ctx context.Context, sourceID string) (*model.Source, error) {
	if sourceID == "" {
		return nil, fmt.Errorf("source ID is required")
	}

	resourcePath, err := resource.SourcePath(sourceID)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s/%s", s.baseURL, resourcePath)

	var source model.Source
	if err := s.doer.DoJSON(ctx, "GET", requestURL, nil, &source); err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	return &source, nil
}
