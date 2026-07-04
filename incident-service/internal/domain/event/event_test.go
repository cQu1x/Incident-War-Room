package event

import "testing"

func TestEventSummary(t *testing.T) {
	tests := []struct {
		name string
		ev   Event
		want string
	}{
		{
			name: "opened event combines label and title",
			ev:   Event{Type: TypeIncidentCreated, Message: "DB is down"},
			want: "Incident opened: DB is down",
		},
		{
			name: "closed event shows its label",
			ev:   Event{Type: TypeIncidentClosed},
			want: "Incident closed",
		},
		{
			name: "reopened event shows its label",
			ev:   Event{Type: TypeIncidentReopened},
			want: "Incident reopened",
		},
		{
			name: "comment shows only its message",
			ev:   Event{Type: TypeCommentAdded, Message: "restarting primary"},
			want: "restarting primary",
		},
		{
			name: "photo comment without caption stays empty",
			ev:   Event{Type: TypeCommentAdded},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ev.Summary(); got != tt.want {
				t.Errorf("Summary() = %q, want %q", got, tt.want)
			}
		})
	}
}
