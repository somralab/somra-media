package plugin

import "errors"

var (
	// ErrIncompatibleContract is returned when a plugin reports an unsupported contract version.
	ErrIncompatibleContract = errors.New("plugin contract version incompatible with core")

	// ErrUnsupportedCapability is returned when an adapter lacks a requested feature.
	ErrUnsupportedCapability = errors.New("plugin capability not supported")

	// ErrPluginNotFound is returned when an instance id does not exist.
	ErrPluginNotFound = errors.New("plugin instance not found")

	// ErrPluginDisabled is returned when a runtime lookup targets a disabled instance.
	ErrPluginDisabled = errors.New("plugin instance disabled")

	// ErrFactoryNotFound is returned when no factory matches the requested implementation.
	ErrFactoryNotFound = errors.New("plugin factory not found")

	// ErrDuplicateFactory is returned when the same factory is registered twice.
	ErrDuplicateFactory = errors.New("plugin factory already registered")

	// ErrDuplicateInstance is returned when an instance name collides within a plugin type.
	ErrDuplicateInstance = errors.New("plugin instance name already exists")
)
