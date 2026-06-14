package requests

import "time"

// Status is the lifecycle state of a content request.
type Status string

const (
	StatusPending   Status = "pending"
	StatusApproved  Status = "approved"
	StatusRejected  Status = "rejected"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

// MediaKind distinguishes movie vs TV show requests.
type MediaKind string

const (
	MediaKindMovie MediaKind = "movie"
	MediaKindTV    MediaKind = "tv"
)

// QualityResolution captures the requester's desired release quality.
// Values match the requests.quality_resolution DB check constraint.
type QualityResolution string

const (
	QualityAny    QualityResolution = "any"
	QualitySD720  QualityResolution = "720p"
	QualityHD1080 QualityResolution = "1080p"
)

// Request is a user-initiated ask for media not yet in the library.
type Request struct {
	ID                int64
	UserID            string
	Status            Status
	MediaKind         MediaKind
	Provider          string
	ExternalID        string
	Title             string
	PosterURL         string
	QualityResolution QualityResolution
	QualityProfile    string
	CollisionFlag     bool
	AdminNote         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
