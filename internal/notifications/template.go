package notifications

import (
	"fmt"

	"golang.org/x/text/language"

	"github.com/somralab/somra-media/internal/platform/i18n"
)

const fallbackLocale = "en-US"

// RenderedMessage holds localized subject and body for an event.
type RenderedMessage struct {
	Subject string
	Body    string
	Locale  string
}

// TemplateRenderer resolves notification copy from the i18n bundle.
type TemplateRenderer struct {
	bundle *i18n.Bundle
}

// NewTemplateRenderer returns a renderer backed by bundle.
func NewTemplateRenderer(bundle *i18n.Bundle) *TemplateRenderer {
	return &TemplateRenderer{bundle: bundle}
}

// Render localizes subject and body for eventType using recipientLocale.
// Missing or unsupported locales fall back to en-US.
func (r *TemplateRenderer) Render(eventType EventType, recipientLocale string, data map[string]any) (RenderedMessage, error) {
	if r == nil || r.bundle == nil {
		return RenderedMessage{}, fmt.Errorf("notifications: template renderer not configured")
	}
	tag := resolveLocaleTag(recipientLocale)
	loc := r.bundle.Localize(tag)

	subjectKey := subjectKeyFor(eventType)
	bodyKey := bodyKeyFor(eventType)
	if subjectKey == "" || bodyKey == "" {
		return RenderedMessage{}, fmt.Errorf("notifications: unknown event type %q", eventType)
	}

	return RenderedMessage{
		Subject: loc.Message(subjectKey, data),
		Body:    loc.Message(bodyKey, data),
		Locale:  tag.String(),
	}, nil
}

func subjectKeyFor(eventType EventType) string {
	switch eventType {
	case EventRequestCreated:
		return "notifications.request.created.subject"
	case EventRequestApproved:
		return "notifications.request.approved.subject"
	case EventRequestRejected:
		return "notifications.request.rejected.subject"
	case EventRequestCompleted:
		return "notifications.request.completed.subject"
	case EventSystemError:
		return "notifications.system.error.subject"
	default:
		return ""
	}
}

func bodyKeyFor(eventType EventType) string {
	switch eventType {
	case EventRequestCreated:
		return "notifications.request.created.body"
	case EventRequestApproved:
		return "notifications.request.approved.body"
	case EventRequestRejected:
		return "notifications.request.rejected.body"
	case EventRequestCompleted:
		return "notifications.request.completed.body"
	case EventSystemError:
		return "notifications.system.error.body"
	default:
		return ""
	}
}

func resolveLocaleTag(locale string) language.Tag {
	switch locale {
	case "tr-TR":
		return language.MustParse("tr-TR")
	case "en-US", "":
		return i18n.SourceLanguage
	default:
		return i18n.SourceLanguage
	}
}
