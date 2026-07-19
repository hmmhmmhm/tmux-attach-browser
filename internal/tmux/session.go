package tmux

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const sessionFormat = "#{session_name}\t#{session_windows}\t#{session_attached}\t#{session_created}\t#{session_activity}"

// ParseSessions parses rows produced with sessionFormat.
func ParseSessions(data []byte) ([]Session, error) {
	input := strings.TrimRight(string(data), "\r\n")
	if input == "" {
		return []Session{}, nil
	}

	lines := strings.Split(input, "\n")
	sessions := make([]Session, 0, len(lines))
	for index, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			return nil, fmt.Errorf("parse session row %d: expected 5 fields, got %d", index+1, len(fields))
		}

		windows, err := parseIntField(fields[1], "window count", index)
		if err != nil {
			return nil, err
		}
		attached, err := parseIntField(fields[2], "attached count", index)
		if err != nil {
			return nil, err
		}
		created, err := parseIntField(fields[3], "created timestamp", index)
		if err != nil {
			return nil, err
		}
		activity, err := parseIntField(fields[4], "activity timestamp", index)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, Session{
			Name:       fields[0],
			Windows:    windows,
			Attached:   attached,
			CreatedAt:  time.Unix(int64(created), 0),
			ActivityAt: time.Unix(int64(activity), 0),
		})
	}

	sort.SliceStable(sessions, func(i, j int) bool {
		if sessions[i].ActivityAt.Equal(sessions[j].ActivityAt) {
			return sessions[i].Name < sessions[j].Name
		}
		return sessions[i].ActivityAt.After(sessions[j].ActivityAt)
	})
	return sessions, nil
}

func parseIntField(value, label string, rowIndex int) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse session row %d %s %q: %w", rowIndex+1, label, value, err)
	}
	return parsed, nil
}

// ValidateSessionName rejects names tmux would rewrite or parse ambiguously.
func ValidateSessionName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("session name cannot be blank")
	}
	if strings.ContainsAny(name, ":.") {
		return fmt.Errorf("session name cannot contain ':' or '.'")
	}
	if strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return fmt.Errorf("session name cannot contain control characters")
	}
	return nil
}
