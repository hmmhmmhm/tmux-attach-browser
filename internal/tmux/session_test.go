package tmux

import (
	"strings"
	"testing"
	"time"
)

func TestParseSessionsEmpty(t *testing.T) {
	sessions, err := ParseSessions(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("len = %d, want 0", len(sessions))
	}
}

func TestParseSessionsSortsRecentActivityFirst(t *testing.T) {
	data := []byte("오래된-세션\t2\t0\t100\t200\nrecent\t3\t1\t300\t500\n")

	sessions, err := ParseSessions(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatalf("len = %d, want 2", len(sessions))
	}
	got := sessions[0]
	if got.Name != "recent" || got.Windows != 3 || got.Attached != 1 {
		t.Fatalf("first session = %#v", got)
	}
	if !got.CreatedAt.Equal(time.Unix(300, 0)) || !got.ActivityAt.Equal(time.Unix(500, 0)) {
		t.Fatalf("timestamps = %v, %v", got.CreatedAt, got.ActivityAt)
	}
	if sessions[1].Name != "오래된-세션" {
		t.Fatalf("second name = %q", sessions[1].Name)
	}
}

func TestParseSessionsUsesNameAsActivityTieBreaker(t *testing.T) {
	sessions, err := ParseSessions([]byte("zeta\t1\t0\t1\t10\nalpha\t1\t0\t1\t10\n"))
	if err != nil {
		t.Fatal(err)
	}
	if sessions[0].Name != "alpha" || sessions[1].Name != "zeta" {
		t.Fatalf("order = %q, %q", sessions[0].Name, sessions[1].Name)
	}
}

func TestParseSessionsRejectsMalformedRows(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{name: "field count", data: "work\t2\t0\n", want: "5 fields"},
		{name: "window count", data: "work\tmany\t0\t1\t2\n", want: "window count"},
		{name: "attached count", data: "work\t1\tmany\t1\t2\n", want: "attached count"},
		{name: "created timestamp", data: "work\t1\t0\tthen\t2\n", want: "created timestamp"},
		{name: "activity timestamp", data: "work\t1\t0\t1\tnow\n", want: "activity timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSessions([]byte(tt.data))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestValidateSessionName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{name: "work", valid: true},
		{name: "프로젝트-1", valid: true},
		{name: "", valid: false},
		{name: "   ", valid: false},
		{name: "a:b", valid: false},
		{name: "a.b", valid: false},
		{name: "a\tb", valid: false},
		{name: "a\nb", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSessionName(tt.name)
			if tt.valid && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
