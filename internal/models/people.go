package models

import "time"

// Person represents a person (face cluster) in Immich.
type Person struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	BirthDate     *string    `json:"birthDate,omitempty"`
	ThumbnailPath string     `json:"thumbnailPath"`
	IsHidden      bool       `json:"isHidden"`
	UpdatedAt     *time.Time `json:"updatedAt,omitempty"`
}

// PeopleResponse is returned by the list people endpoint.
type PeopleResponse struct {
	People  []PersonWithFaces `json:"people"`
	Total   int               `json:"total"`
	Visible int               `json:"visible"`
	Hidden  int               `json:"hidden"`
}

// PersonWithFaces extends Person with face data.
type PersonWithFaces struct {
	Person
	Faces []FaceInfo `json:"faces,omitempty"`
}

// FaceInfo holds bounding box data for a face.
type FaceInfo struct {
	ID           string `json:"id"`
	ImageHeight  int    `json:"imageHeight"`
	ImageWidth   int    `json:"imageWidth"`
	BoundingBoxX1 int   `json:"boundingBoxX1"`
	BoundingBoxX2 int   `json:"boundingBoxX2"`
	BoundingBoxY1 int   `json:"boundingBoxY1"`
	BoundingBoxY2 int   `json:"boundingBoxY2"`
}

// PersonUpdateRequest is used to update a person.
type PersonUpdateRequest struct {
	Name               *string `json:"name,omitempty"`
	BirthDate          *string `json:"birthDate,omitempty"`
	IsHidden           *bool   `json:"isHidden,omitempty"`
	FeatureFaceAssetID *string `json:"featureFaceAssetId,omitempty"`
}

// PersonSummary is a lightweight representation of a person.
type PersonSummary struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	BirthDate     *string `json:"birthDate,omitempty"`
	IsHidden      bool    `json:"isHidden"`
	ThumbnailPath string  `json:"thumbnailPath"`
}

// PersonSummaryFromPerson converts a Person to a PersonSummary.
func PersonSummaryFromPerson(p Person) PersonSummary {
	return PersonSummary{
		ID:            p.ID,
		Name:          p.Name,
		BirthDate:     p.BirthDate,
		IsHidden:      p.IsHidden,
		ThumbnailPath: p.ThumbnailPath,
	}
}
