package server

import (
	"testing"

	"github.com/charmbracelet/crush/internal/session"
	"github.com/stretchr/testify/require"
)

func TestParseServerGoalStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want session.GoalStatus
		ok   bool
	}{
		{in: "active", want: session.GoalStatusActive, ok: true},
		{in: "paused", want: session.GoalStatusPaused, ok: true},
		{in: "blocked", want: session.GoalStatusBlocked, ok: true},
		{in: "usageLimited", want: session.GoalStatusUsageLimited, ok: true},
		{in: "budgetLimited", want: session.GoalStatusBudgetLimited, ok: true},
		{in: "complete", want: session.GoalStatusComplete, ok: true},
		{in: "unknown", want: "", ok: false},
	}

	for _, tc := range tests {
		got, ok := parseServerGoalStatus(tc.in)
		require.Equal(t, tc.ok, ok, tc.in)
		require.Equal(t, tc.want, got, tc.in)
	}
}
