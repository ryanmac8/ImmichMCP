package models

import "time"

// Activity represents a comment or like in Immich.
type Activity struct {
	ID        string       `json:"id"`
	CreatedAt time.Time    `json:"createdAt"`
	Type      string       `json:"type"`
	Comment   *string      `json:"comment,omitempty"`
	AssetID   *string      `json:"assetId,omitempty"`
	User      ActivityUser `json:"user"`
}

// ActivityUser holds user information for an activity.
type ActivityUser struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	ProfileImagePath string `json:"profileImagePath"`
}

// ActivityCreateRequest is used to create an activity.
type ActivityCreateRequest struct {
	AlbumID string  `json:"albumId"`
	AssetID *string `json:"assetId,omitempty"`
	Type    string  `json:"type"`
	Comment *string `json:"comment,omitempty"`
}

// ActivityStatistics holds activity counts.
type ActivityStatistics struct {
	Comments int `json:"comments"`
}

// ActivitySummary is a lightweight representation of an activity.
type ActivitySummary struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Type      string    `json:"type"`
	Comment   *string   `json:"comment,omitempty"`
	AssetID   *string   `json:"assetId,omitempty"`
	UserName  string    `json:"userName"`
}

// ActivitySummaryFromActivity converts an Activity to an ActivitySummary.
func ActivitySummaryFromActivity(a Activity) ActivitySummary {
	return ActivitySummary{
		ID:        a.ID,
		CreatedAt: a.CreatedAt,
		Type:      a.Type,
		Comment:   a.Comment,
		AssetID:   a.AssetID,
		UserName:  a.User.Name,
	}
}
