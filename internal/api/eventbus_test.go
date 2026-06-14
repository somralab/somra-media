package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/library"
)

func TestScanProgressPublisher_Publishes(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe()
	pub := ScanProgressPublisher{Bus: bus}
	pub.PublishScanProgress(context.Background(), library.ProgressEvent{
		LibraryID: 1, ScanRunID: 2, FilesDone: 3, FilesTotal: 10, Status: "running",
	})
	select {
	case msg := <-ch:
		require.Contains(t, string(msg), "scan.progress")
	default:
		t.Fatal("expected scan.progress event")
	}
}

func TestScanProgressPublisher_NilBus(t *testing.T) {
	pub := ScanProgressPublisher{}
	pub.PublishScanProgress(context.Background(), library.ProgressEvent{})
}
