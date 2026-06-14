package notifications_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/i18n"
)

func TestTemplateRendererEnglish(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)
	r := notifications.NewTemplateRenderer(b)

	msg, err := r.Render(notifications.EventRequestCreated, "en-US", map[string]any{
		"Title":  "Inception",
		"Detail": "Blu-ray",
	})
	require.NoError(t, err)
	assert.Contains(t, msg.Subject, "Inception")
	assert.Contains(t, msg.Body, "Blu-ray")
	assert.Equal(t, "en-US", msg.Locale)
}

func TestTemplateRendererTurkish(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)
	r := notifications.NewTemplateRenderer(b)

	msg, err := r.Render(notifications.EventRequestRejected, "tr-TR", map[string]any{
		"Title":  "Dune",
		"Detail": "Red gerekçesi",
	})
	require.NoError(t, err)
	assert.Contains(t, msg.Subject, "reddedildi")
	assert.Equal(t, "tr-TR", msg.Locale)
}

func TestTemplateRendererFallbackUnknownLocale(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)
	r := notifications.NewTemplateRenderer(b)

	msg, err := r.Render(notifications.EventRequestCompleted, "fr-FR", map[string]any{"Title": "X"})
	require.NoError(t, err)
	assert.Contains(t, msg.Subject, "completed")
	assert.Equal(t, "en-US", msg.Locale)
}

func TestTemplateRendererUnknownEvent(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)
	r := notifications.NewTemplateRenderer(b)

	_, err = r.Render("unknown.event", "en-US", nil)
	require.Error(t, err)
}
