package models

import "time"

// SharedLink represents a shared link in Immich.
type SharedLink struct {
	ID           string     `json:"id"`
	Key          string     `json:"key"`
	Type         string     `json:"type"`
	CreatedAt    time.Time  `json:"createdAt"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
	UserID       string     `json:"userId"`
	AllowUpload  bool       `json:"allowUpload"`
	AllowDownload bool      `json:"allowDownload"`
	ShowMetadata bool       `json:"showMetadata"`
	Password     *string    `json:"password,omitempty"`
	Description  *string    `json:"description,omitempty"`
	Album        *Album     `json:"album,omitempty"`
	Assets       []Asset    `json:"assets,omitempty"`
}

// SharedLinkCreateRequest is used to create a shared link.
type SharedLinkCreateRequest struct {
	Type          string     `json:"type"`
	AlbumID       *string    `json:"albumId,omitempty"`
	AssetIDs      []string   `json:"assetIds,omitempty"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	AllowUpload   *bool      `json:"allowUpload,omitempty"`
	AllowDownload *bool      `json:"allowDownload,omitempty"`
	ShowMetadata  *bool      `json:"showMetadata,omitempty"`
	Password      *string    `json:"password,omitempty"`
	Description   *string    `json:"description,omitempty"`
}

// SharedLinkUpdateRequest is used to update a shared link.
type SharedLinkUpdateRequest struct {
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	AllowUpload    *bool      `json:"allowUpload,omitempty"`
	AllowDownload  *bool      `json:"allowDownload,omitempty"`
	ShowMetadata   *bool      `json:"showMetadata,omitempty"`
	Password       *string    `json:"password,omitempty"`
	Description    *string    `json:"description,omitempty"`
	ChangeExpiryTime *bool    `json:"changeExpiryTime,omitempty"`
}

// SharedLinkSummary is a lightweight representation of a shared link.
type SharedLinkSummary struct {
	ID            string     `json:"id"`
	Key           string     `json:"key"`
	Type          string     `json:"type"`
	CreatedAt     time.Time  `json:"createdAt"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	AllowUpload   bool       `json:"allowUpload"`
	AllowDownload bool       `json:"allowDownload"`
	ShowMetadata  bool       `json:"showMetadata"`
	Description   *string    `json:"description,omitempty"`
	AlbumName     *string    `json:"album_name,omitempty"`
	AssetCount    int        `json:"asset_count"`
}

// SharedLinkSummaryFromLink converts a SharedLink to a SharedLinkSummary.
func SharedLinkSummaryFromLink(l SharedLink) SharedLinkSummary {
	s := SharedLinkSummary{
		ID:            l.ID,
		Key:           l.Key,
		Type:          l.Type,
		CreatedAt:     l.CreatedAt,
		ExpiresAt:     l.ExpiresAt,
		AllowUpload:   l.AllowUpload,
		AllowDownload: l.AllowDownload,
		ShowMetadata:  l.ShowMetadata,
		Description:   l.Description,
		AssetCount:    len(l.Assets),
	}
	if l.Album != nil {
		s.AlbumName = &l.Album.AlbumName
		if len(l.Assets) == 0 {
			s.AssetCount = l.Album.AssetCount
		}
	}
	return s
}
