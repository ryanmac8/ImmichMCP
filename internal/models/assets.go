package models

import "time"

// Asset represents a photo or video in Immich.
type Asset struct {
	ID               string      `json:"id"`
	DeviceAssetID    string      `json:"deviceAssetId"`
	OwnerID          string      `json:"ownerId"`
	DeviceID         string      `json:"deviceId"`
	LibraryID        *string     `json:"libraryId,omitempty"`
	Type             string      `json:"type"`
	OriginalPath     string      `json:"originalPath"`
	OriginalFileName string      `json:"originalFileName"`
	OriginalMimeType *string     `json:"originalMimeType,omitempty"`
	Thumbhash        *string     `json:"thumbhash,omitempty"`
	FileCreatedAt    time.Time   `json:"fileCreatedAt"`
	FileModifiedAt   time.Time   `json:"fileModifiedAt"`
	LocalDateTime    time.Time   `json:"localDateTime"`
	UpdatedAt        time.Time   `json:"updatedAt"`
	IsFavorite       bool        `json:"isFavorite"`
	IsArchived       bool        `json:"isArchived"`
	IsTrashed        bool        `json:"isTrashed"`
	IsOffline        bool        `json:"isOffline"`
	HasMetadata      bool        `json:"hasMetadata"`
	Duration         string      `json:"duration"`
	ExifInfo         *ExifInfo   `json:"exifInfo,omitempty"`
	LivePhotoVideoID *string     `json:"livePhotoVideoId,omitempty"`
	People           []PersonFace `json:"people,omitempty"`
	Checksum         string      `json:"checksum"`
	StackCount       *int        `json:"stackCount,omitempty"`
	Stack            *AssetStack `json:"stack,omitempty"`
	DuplicateID      *string     `json:"duplicateId,omitempty"`
	Resized          bool        `json:"resized"`
}

// ExifInfo holds EXIF metadata for an asset.
type ExifInfo struct {
	Make             *string    `json:"make,omitempty"`
	Model            *string    `json:"model,omitempty"`
	ExifImageWidth   *int       `json:"exifImageWidth,omitempty"`
	ExifImageHeight  *int       `json:"exifImageHeight,omitempty"`
	FileSizeInByte   *int64     `json:"fileSizeInByte,omitempty"`
	Orientation      *string    `json:"orientation,omitempty"`
	DateTimeOriginal *time.Time `json:"dateTimeOriginal,omitempty"`
	ModifyDate       *time.Time `json:"modifyDate,omitempty"`
	TimeZone         *string    `json:"timeZone,omitempty"`
	LensModel        *string    `json:"lensModel,omitempty"`
	FNumber          *float64   `json:"fNumber,omitempty"`
	FocalLength      *float64   `json:"focalLength,omitempty"`
	ISO              *int       `json:"iso,omitempty"`
	ExposureTime     *string    `json:"exposureTime,omitempty"`
	Latitude         *float64   `json:"latitude,omitempty"`
	Longitude        *float64   `json:"longitude,omitempty"`
	City             *string    `json:"city,omitempty"`
	State            *string    `json:"state,omitempty"`
	Country          *string    `json:"country,omitempty"`
	Description      *string    `json:"description,omitempty"`
	ProjectionType   *string    `json:"projectionType,omitempty"`
	Rating           *int       `json:"rating,omitempty"`
}

// PersonFace is face information attached to an asset.
type PersonFace struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	BirthDate     *string `json:"birthDate,omitempty"`
	ThumbnailPath *string `json:"thumbnailPath,omitempty"`
	IsHidden      bool    `json:"isHidden"`
}

// AssetStack holds stack information for an asset.
type AssetStack struct {
	ID             string `json:"id"`
	PrimaryAssetID string `json:"primaryAssetId"`
	AssetCount     int    `json:"assetCount"`
}

// AssetStatistics holds asset counts.
type AssetStatistics struct {
	Images int `json:"images"`
	Videos int `json:"videos"`
	Total  int `json:"total"`
}

// AssetUpdateRequest is used to update an asset.
type AssetUpdateRequest struct {
	IsFavorite       *bool      `json:"isFavorite,omitempty"`
	IsArchived       *bool      `json:"isArchived,omitempty"`
	Description      *string    `json:"description,omitempty"`
	DateTimeOriginal *time.Time `json:"dateTimeOriginal,omitempty"`
	Latitude         *float64   `json:"latitude,omitempty"`
	Longitude        *float64   `json:"longitude,omitempty"`
	Rating           *int       `json:"rating,omitempty"`
}

// AssetBulkUpdateRequest is used to bulk-update assets.
type AssetBulkUpdateRequest struct {
	IDs              []string   `json:"ids"`
	IsFavorite       *bool      `json:"isFavorite,omitempty"`
	IsArchived       *bool      `json:"isArchived,omitempty"`
	DateTimeOriginal *time.Time `json:"dateTimeOriginal,omitempty"`
	Latitude         *float64   `json:"latitude,omitempty"`
	Longitude        *float64   `json:"longitude,omitempty"`
	DuplicateID      *string    `json:"duplicateId,omitempty"`
	Rating           *int       `json:"rating,omitempty"`
}

// AssetDownloadInfo holds URLs for downloading an asset.
type AssetDownloadInfo struct {
	ID               string  `json:"id"`
	OriginalFileName *string `json:"original_file_name,omitempty"`
	OriginalURL      string  `json:"original_url"`
	ThumbnailURL     *string `json:"thumbnail_url,omitempty"`
	PreviewURL       *string `json:"preview_url,omitempty"`
}

// AssetSummary is a lightweight representation of an asset.
type AssetSummary struct {
	ID               string     `json:"id"`
	Type             string     `json:"type"`
	OriginalFileName string     `json:"originalFileName"`
	FileCreatedAt    time.Time  `json:"fileCreatedAt"`
	LocalDateTime    time.Time  `json:"localDateTime"`
	IsFavorite       bool       `json:"isFavorite"`
	IsArchived       bool       `json:"isArchived"`
	Duration         string     `json:"duration"`
	City             *string    `json:"city,omitempty"`
	Country          *string    `json:"country,omitempty"`
	Make             *string    `json:"make,omitempty"`
	Model            *string    `json:"model,omitempty"`
	Thumbhash        *string    `json:"thumbhash,omitempty"`
}

// AssetSummaryFromAsset converts an Asset to an AssetSummary.
func AssetSummaryFromAsset(a Asset) AssetSummary {
	s := AssetSummary{
		ID:               a.ID,
		Type:             a.Type,
		OriginalFileName: a.OriginalFileName,
		FileCreatedAt:    a.FileCreatedAt,
		LocalDateTime:    a.LocalDateTime,
		IsFavorite:       a.IsFavorite,
		IsArchived:       a.IsArchived,
		Duration:         a.Duration,
		Thumbhash:        a.Thumbhash,
	}
	if a.ExifInfo != nil {
		s.City = a.ExifInfo.City
		s.Country = a.ExifInfo.Country
		s.Make = a.ExifInfo.Make
		s.Model = a.ExifInfo.Model
	}
	return s
}
