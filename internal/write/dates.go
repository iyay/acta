package write

import "github.com/iyay/acta/internal/board"

// The board works most statuses out from other data, so a date is written
// by the command that changes the file, never by the agent.

// Working says whether a status means someone has begun the work.
func Working(status string) bool {
	switch status {
	case "in-progress", "fixing", "brainstorming":
		return true
	}
	return false
}

// MarkStarted writes today as the day the work began, but only the first
// time: a file that already has a started keeps it.
func MarkStarted(src []byte) ([]byte, error) {
	if hasField(src, "started") {
		return src, nil
	}
	return SetField(src, "started", Now().Format("2006-01-02"))
}

// MarkFinished writes today as the day the work closed. An old finished is
// replaced, so closing again counts from today.
func MarkFinished(src []byte) ([]byte, error) {
	return SetField(src, "finished", Now().Format("2006-01-02"))
}

// ClearFinished takes the close day away when the work opens again.
func ClearFinished(src []byte) ([]byte, error) { return RemoveField(src, "finished") }

// DatesFor gives the file the dates its new status calls for.
func DatesFor(src []byte, status string) ([]byte, error) {
	switch {
	case Working(status):
		out, err := MarkStarted(src)
		if err != nil {
			return nil, err
		}
		return ClearFinished(out)
	case board.Closed(status):
		return MarkFinished(src)
	}
	return src, nil
}
