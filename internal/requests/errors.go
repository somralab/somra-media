package requests

import "errors"

var (
	// ErrInvalidTransition is returned when a status change is not allowed.
	ErrInvalidTransition = errors.New("requests: invalid status transition")

	// ErrCollisionInLibrary is returned when the item already exists in the library.
	ErrCollisionInLibrary = errors.New("requests: item already in library")

	// ErrCollisionDuplicatePending is returned when an identical pending request exists.
	ErrCollisionDuplicatePending = errors.New("requests: duplicate pending request")

	// ErrQuotaExceeded is returned when the user has exhausted their monthly quota.
	ErrQuotaExceeded = errors.New("requests: monthly quota exceeded")
)
