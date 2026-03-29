package models

import "time"

// Tag represents a tag in Immich.
type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Color     *string   `json:"color,omitempty"`
}

// TagCreateRequest is used to create a tag.
type TagCreateRequest struct {
	Name  string  `json:"name"`
	Color *string `json:"color,omitempty"`
}

// TagUpdateRequest is used to update a tag.
type TagUpdateRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

// TagSummary is a lightweight representation of a tag.
type TagSummary struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Value string  `json:"value"`
	Color *string `json:"color,omitempty"`
}

// TagSummaryFromTag converts a Tag to a TagSummary.
func TagSummaryFromTag(t Tag) TagSummary {
	return TagSummary{
		ID:    t.ID,
		Name:  t.Name,
		Value: t.Value,
		Color: t.Color,
	}
}
