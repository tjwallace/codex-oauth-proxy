package logger

import (
	"encoding/json"
	"io"
)

// privacyWriter permits only operational fields. Upstream logs include prompts,
// tool arguments, OAuth token fragments and upstream error bodies even at info/warn.
// Unknown fields and malformed records fail closed, without printing their values.
type privacyWriter struct{ output io.Writer }

func (w privacyWriter) Write(p []byte) (int, error) {
	var event map[string]json.RawMessage
	if err := json.Unmarshal(p, &event); err != nil {
		return len(p), nil
	}
	allowed := map[string]bool{
		"time": true, "level": true, "message": true,
		"status": true, "status_code": true, "duration": true,
		"duration_ms": true, "method": true, "input_count": true,
		"minutes_until_expiry": true, "new_expiry_minutes": true,
	}
	for key := range event {
		if !allowed[key] {
			delete(event, key)
		}
	}
	data, err := json.Marshal(event)
	if err != nil {
		return len(p), err
	}
	_, err = w.output.Write(append(data, '\n'))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
