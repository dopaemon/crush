package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/crush/internal/proto"
	"github.com/stretchr/testify/require"
)

func TestThreadGoalGet(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/v1/workspaces/ws1/thread/goal/get", r.URL.Path)
		require.Equal(t, "sess1", r.URL.Query().Get("session_id"))
		_ = json.NewEncoder(w).Encode(proto.ThreadGoalGetResponse{
			Goal: &proto.Goal{Objective: "ship", Status: "active"},
		})
	}))
	defer srv.Close()

	c := captureClient(t, srv)
	resp, err := c.ThreadGoalGet(context.Background(), "ws1", "sess1")
	require.NoError(t, err)
	require.NotNil(t, resp.Goal)
	require.Equal(t, "ship", resp.Goal.Objective)
	require.Equal(t, "active", resp.Goal.Status)
}

func TestThreadGoalSet(t *testing.T) {
	t.Parallel()

	var got proto.ThreadGoalSetRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/workspaces/ws1/thread/goal/set", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &got))
		_ = json.NewEncoder(w).Encode(proto.ThreadGoalSetResponse{
			Goal: &proto.Goal{Objective: got.Objective, Status: "active"},
		})
	}))
	defer srv.Close()

	c := captureClient(t, srv)
	status := "active"
	resp, err := c.ThreadGoalSet(context.Background(), "ws1", proto.ThreadGoalSetRequest{
		SessionID: "sess1",
		Objective: "finish parity",
		Status:    &status,
	})
	require.NoError(t, err)
	require.Equal(t, "sess1", got.SessionID)
	require.Equal(t, "finish parity", got.Objective)
	require.NotNil(t, got.Status)
	require.Equal(t, "active", *got.Status)
	require.NotNil(t, resp.Goal)
	require.Equal(t, "finish parity", resp.Goal.Objective)
}

func TestThreadGoalClear(t *testing.T) {
	t.Parallel()

	var got proto.ThreadGoalClearRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/workspaces/ws1/thread/goal/clear", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &got))
		_ = json.NewEncoder(w).Encode(proto.ThreadGoalClearResponse{Cleared: true})
	}))
	defer srv.Close()

	c := captureClient(t, srv)
	resp, err := c.ThreadGoalClear(context.Background(), "ws1", proto.ThreadGoalClearRequest{
		SessionID: "sess1",
	})
	require.NoError(t, err)
	require.Equal(t, "sess1", got.SessionID)
	require.True(t, resp.Cleared)
}
