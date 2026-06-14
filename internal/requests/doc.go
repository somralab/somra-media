// Package requests implements content request lifecycle for Somra Sprint 08:
// model types, controlled status transitions, library/request collision checks,
// quota and auto-approve policy evaluation, TMDB-backed discover search, and
// the automation handoff seam for Sprint 09.
//
// Wave 2 (HTTP handlers) and Wave 1A (migrations + request repository CRUD)
// wire into this package via the Repository and PolicyStore interfaces.
//
// # Automation handoff (Sprint 09)
//
// Approved requests are handed off to the automation module through
// [AutomationHandoff]. The default [NoOpAutomationHandoff] satisfies the
// interface without side effects. [QueuingAutomationHandoff] optionally records
// handoff intent via [HandoffQueue] so Sprint 09 can drain a queue without
// changing call sites in the request service.
package requests
