package write

import "github.com/iyay/acta/internal/board"

// The board works most statuses out from other data, so a date is written
// by the command that changes the file, never by the agent.

// stampLayout is how a date field is written: the day and the time to the
// second, so the reader can tell two changes on one day apart.
const stampLayout = "2006-01-02 15:04:05"

// Working says whether a status means someone has begun the work.
func Working(status string) bool {
	switch status {
	case "in-progress", "fixing", "brainstorming":
		return true
	}
	return false
}

// MarkStarted writes now as the moment the work began, but only the first
// time: a file that already has a started keeps it.
func MarkStarted(src []byte) ([]byte, error) {
	if hasField(src, "started") {
		return src, nil
	}
	return SetField(src, "started", Now().Format(stampLayout))
}

// MarkFinished writes now as the moment the work closed. An old finished is
// replaced, so closing again counts from now.
func MarkFinished(src []byte) ([]byte, error) {
	return SetField(src, "finished", Now().Format(stampLayout))
}

// MarkFinishedOnce writes now as the moment the work closed, but only the
// first time. An automatic close, like a tick or a spec closing the idea it
// came from, must not move the day a person already read.
func MarkFinishedOnce(src []byte) ([]byte, error) {
	if hasField(src, "finished") {
		return src, nil
	}
	return MarkFinished(src)
}

// ClearFinished takes the close moment away when the work opens again.
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
