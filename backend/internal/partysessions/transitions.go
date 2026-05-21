package partysessions

// validStatus is the closed set of party-session statuses defined by the
// OpenAPI contract and the CHECK constraint on party_sessions.status.
func validStatus(s string) bool {
	switch s {
	case "active", "paused", "ended":
		return true
	}
	return false
}

// legalTransition reports whether moving a party session from status `from`
// to status `to` is allowed.
//
//   - active  → active, paused, ended
//   - paused  → paused, active, ended
//   - ended   → ended (terminal)
//
// Same-value transitions are allowed; the UPDATE still runs and ticks
// updated_at, consistent with the rest of the API's PATCH semantics.
func legalTransition(from, to string) bool {
	if !validStatus(from) || !validStatus(to) {
		return false
	}
	if from == "ended" {
		return to == "ended"
	}
	return true
}
