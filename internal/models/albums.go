package models

import "time"

// Album represents an album in Immich.
type Album struct {
	ID                       string       `json:"id"`
	OwnerID                  string       `json:"ownerId"`
	AlbumName                string       `json:"albumName"`
	Description              string       `json:"description"`
	CreatedAt                time.Time    `json:"createdAt"`
	UpdatedAt                time.Time    `json:"updatedAt"`
	AlbumThumbnailAssetID    *string      `json:"albumThumbnailAssetId,omitempty"`
	Shared                   bool         `json:"shared"`
	HasSharedLink            bool         `json:"hasSharedLink"`
	StartDate                *time.Time   `json:"startDate,omitempty"`
	EndDate                  *time.Time   `json:"endDate,omitempty"`
	Assets                   []Asset      `json:"assets,omitempty"`
	AssetCount               int          `json:"assetCount"`
	Owner                    *AlbumOwner  `json:"owner,omitempty"`
	SharedUsers              []AlbumUser  `json:"sharedUsers,omitempty"`
	IsActivityEnabled        bool         `json:"isActivityEnabled"`
	Order                    *string      `json:"order,omitempty"`
	LastModifiedAssetTimestamp *time.Time `json:"lastModifiedAssetTimestamp,omitempty"`
}

// AlbumOwner holds owner information for an album.
type AlbumOwner struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	ProfileImagePath string `json:"profileImagePath"`
}

// AlbumUser holds shared-user information for an album.
type AlbumUser struct {
	User AlbumOwner `json:"user"`
	Role string     `json:"role"`
}

// AlbumCreateRequest is used to create an album.
type AlbumCreateRequest struct {
	AlbumName          string   `json:"albumName"`
	Description        *string  `json:"description,omitempty"`
	AssetIDs           []string `json:"assetIds,omitempty"`
	SharedWithUserIDs  []string `json:"sharedWithUserIds,omitempty"`
}

// AlbumUpdateRequest is used to update an album.
type AlbumUpdateRequest struct {
	AlbumName              *string `json:"albumName,omitempty"`
	Description            *string `json:"description,omitempty"`
	AlbumThumbnailAssetID  *string `json:"albumThumbnailAssetId,omitempty"`
	IsActivityEnabled      *bool   `json:"isActivityEnabled,omitempty"`
	Order                  *string `json:"order,omitempty"`
}

// AlbumStatistics holds album counts.
type AlbumStatistics struct {
	Owned     int `json:"owned"`
	Shared    int `json:"shared"`
	NotShared int `json:"notShared"`
}

// AlbumSummary is a lightweight representation of an album.
type AlbumSummary struct {
	ID                    string     `json:"id"`
	AlbumName             string     `json:"albumName"`
	Description           string     `json:"description"`
	AssetCount            int        `json:"assetCount"`
	Shared                bool       `json:"shared"`
	StartDate             *time.Time `json:"startDate,omitempty"`
	EndDate               *time.Time `json:"endDate,omitempty"`
	AlbumThumbnailAssetID *string    `json:"albumThumbnailAssetId,omitempty"`
}

// AlbumSummaryFromAlbum converts an Album to an AlbumSummary.
func AlbumSummaryFromAlbum(a Album) AlbumSummary {
	return AlbumSummary{
		ID:                    a.ID,
		AlbumName:             a.AlbumName,
		Description:           a.Description,
		AssetCount:            a.AssetCount,
		Shared:                a.Shared,
		StartDate:             a.StartDate,
		EndDate:               a.EndDate,
		AlbumThumbnailAssetID: a.AlbumThumbnailAssetID,
	}
}
