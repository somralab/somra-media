package plugin

import "errors"

var (
	// ErrIncompatibleContract is returned when a plugin reports an unsupported contract version.
	ErrIncompatibleContract = errors.New("plugin contract version incompatible with core")

	// ErrUnsupportedCapability is returned when an adapter lacks a requested feature.
	ErrUnsupportedCapability = errors.New("plugin capability not supported")
)
