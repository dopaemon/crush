package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/charmbracelet/crush/internal/proto"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/stretchr/testify/require"
)

func TestSubscribeEvents_ParsesGoalNotification(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/workspaces/ws1/events", r.URL.Path)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		inner, err := json.Marshal(pubsub.Event[proto.GoalNotification]{
			Type: pubsub.UpdatedEvent,
			Payload: proto.GoalNotification{
				SessionID: "sess1",
				Method:    "thread/goal/updated",
				Goal:      &proto.Goal{Objective: "parity", Status: "active"},
			},
		})
		require.NoError(t, err)
		env, err := json.Marshal(pubsub.Payload{
			Type:    pubsub.PayloadTypeGoalNotification,
			Payload: inner,
		})
		require.NoError(t, err)

		fmt.Fprintf(w, "data: %s\n\n", env)
		flusher.Flush()
		time.Sleep(50 * time.Millisecond)
	}))
	defer srv.Close()

	c := captureClient(t, srv)
	events, err := c.SubscribeEvents(context.Background(), "ws1")
	require.NoError(t, err)

	select {
	case ev := <-events:
		msg, ok := ev.(pubsub.Event[proto.GoalNotification])
		require.True(t, ok, "expected goal notification event, got %T", ev)
		require.Equal(t, "sess1", msg.Payload.SessionID)
		require.Equal(t, "thread/goal/updated", msg.Payload.Method)
		require.NotNil(t, msg.Payload.Goal)
		require.Equal(t, "parity", msg.Payload.Goal.Objective)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for goal notification event")
	}
}
