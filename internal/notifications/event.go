package notifications

// Event carries domain data for notification rendering and routing.
type Event struct {
	Type     EventType
	UserID   string
	AdminIDs []string
	Title    string
	Detail   string
	ErrorMsg string
	Data     map[string]any
}

// TemplateData merges explicit fields with Event.Data for i18n templates.
func (e Event) TemplateData() map[string]any {
	out := map[string]any{
		"Title":  e.Title,
		"Detail": e.Detail,
		"Error":  e.ErrorMsg,
	}
	for k, v := range e.Data {
		out[k] = v
	}
	return out
}

// Recipients returns user IDs that should receive this event.
func (e Event) Recipients() []string {
	seen := make(map[string]struct{}, 1+len(e.AdminIDs))
	out := make([]string, 0, 1+len(e.AdminIDs))
	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	switch e.Type {
	case EventRequestCreated, EventRequestApproved, EventRequestRejected, EventRequestCompleted:
		add(e.UserID)
		for _, id := range e.AdminIDs {
			add(id)
		}
	case EventSystemError:
		for _, id := range e.AdminIDs {
			add(id)
		}
	}
	return out
}
