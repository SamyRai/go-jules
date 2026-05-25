package jules

import (
	"context"
	"fmt"
	"net/url"
)

// ListSourcesOptions controls source pagination and filtering.
type ListSourcesOptions struct {
	PageSize  int
	PageToken string
	Filter    string
}

// SourcesService owns Jules source operations.
type SourcesService struct {
	transport *transport
}

// List lists available code sources with pagination and filtering.
func (s *SourcesService) List(ctx context.Context, options *ListSourcesOptions) (*SourcesResponse, error) {
	pageSize := 30
	pageToken := ""
	filter := ""
	if options != nil {
		pageSize = options.PageSize
		pageToken = options.PageToken
		filter = options.Filter
	}
	if pageSize <= 0 {
		pageSize = 30 // default page size per API docs
	}
	if pageSize > 100 {
		pageSize = 100 // max page size per API docs
	}

	query := url.Values{}
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}
	if filter != "" {
		query.Set("filter", filter)
	}
	requestURL := fmt.Sprintf("%s/sources?%s", s.transport.client.BaseURL, query.Encode())

	var response SourcesResponse
	if err := s.transport.doJSON(ctx, "GET", requestURL, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}

	return &response, nil
}

// ListAll retrieves every source by following nextPageToken.
func (s *SourcesService) ListAll(ctx context.Context, pageSize int, filter string) ([]Source, error) {
	var sources []Source
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
func (s *SourcesService) Get(ctx context.Context, sourceID string) (*Source, error) {
	if sourceID == "" {
		return nil, fmt.Errorf("source ID is required")
	}

	resourcePath, err := sourcePath(sourceID)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s/%s", s.transport.client.BaseURL, resourcePath)

	var source Source
	if err := s.transport.doJSON(ctx, "GET", requestURL, nil, &source); err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	return &source, nil
}
