package models

import "time"

// MetadataSearchRequest is the request body for metadata search.
type MetadataSearchRequest struct {
	Page             *int       `json:"page,omitempty"`
	Size             *int       `json:"size,omitempty"`
	Type             *string    `json:"type,omitempty"`
	IsFavorite       *bool      `json:"isFavorite,omitempty"`
	IsArchived       *bool      `json:"isArchived,omitempty"`
	IsTrashed        *bool      `json:"isTrashed,omitempty"`
	IsVisible        *bool      `json:"isVisible,omitempty"`
	IsMotion         *bool      `json:"isMotion,omitempty"`
	IsNotInAlbum     *bool      `json:"isNotInAlbum,omitempty"`
	IsOffline        *bool      `json:"isOffline,omitempty"`
	WithExif         *bool      `json:"withExif,omitempty"`
	WithPeople       *bool      `json:"withPeople,omitempty"`
	TakenAfter       *time.Time `json:"takenAfter,omitempty"`
	TakenBefore      *time.Time `json:"takenBefore,omitempty"`
	UpdatedAfter     *time.Time `json:"updatedAfter,omitempty"`
	UpdatedBefore    *time.Time `json:"updatedBefore,omitempty"`
	City             *string    `json:"city,omitempty"`
	State            *string    `json:"state,omitempty"`
	Country          *string    `json:"country,omitempty"`
	Make             *string    `json:"make,omitempty"`
	Model            *string    `json:"model,omitempty"`
	LensModel        *string    `json:"lensModel,omitempty"`
	PersonIDs        []string   `json:"personIds,omitempty"`
	OriginalFileName *string    `json:"originalFileName,omitempty"`
	OriginalPath     *string    `json:"originalPath,omitempty"`
	Order            *string    `json:"order,omitempty"`
}

// SmartSearchRequest is the request body for smart/CLIP search.
type SmartSearchRequest struct {
	Query        string     `json:"query"`
	Page         *int       `json:"page,omitempty"`
	Size         *int       `json:"size,omitempty"`
	Type         *string    `json:"type,omitempty"`
	IsFavorite   *bool      `json:"isFavorite,omitempty"`
	IsArchived   *bool      `json:"isArchived,omitempty"`
	IsTrashed    *bool      `json:"isTrashed,omitempty"`
	IsVisible    *bool      `json:"isVisible,omitempty"`
	City         *string    `json:"city,omitempty"`
	State        *string    `json:"state,omitempty"`
	Country      *string    `json:"country,omitempty"`
	Make         *string    `json:"make,omitempty"`
	Model        *string    `json:"model,omitempty"`
	TakenAfter   *time.Time `json:"takenAfter,omitempty"`
	TakenBefore  *time.Time `json:"takenBefore,omitempty"`
	PersonIDs    []string   `json:"personIds,omitempty"`
}

// SearchResult is the top-level response from the search API.
type SearchResult struct {
	Assets SearchAssetResult `json:"assets"`
}

// SearchAssetResult holds the paginated search results.
type SearchAssetResult struct {
	Count    int     `json:"count"`
	Total    int     `json:"total"`
	Items    []Asset `json:"items"`
	NextPage *string `json:"nextPage,omitempty"`
}

// ExploreData is returned by the explore endpoint.
type ExploreData struct {
	FieldName string        `json:"fieldName"`
	Items     []ExploreItem `json:"items"`
}

// ExploreItem is a single explore result.
type ExploreItem struct {
	Value string          `json:"value"`
	Data  ExploreItemData `json:"data"`
}

// ExploreItemData holds data for an explore item.
type ExploreItemData struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Thumbhash *string `json:"thumbhash,omitempty"`
}
