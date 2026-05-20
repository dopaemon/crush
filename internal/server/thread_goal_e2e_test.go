package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/proto"
	"github.com/stretchr/testify/require"
)

func TestThreadGoalEndpointsE2E(t *testing.T) {
	t.Parallel()

	wd := t.TempDir()
	cfg, err := config.Init(wd, "", false)
	require.NoError(t, err)

	s := NewServer(cfg, "tcp", "127.0.0.1:0")
	ts := httptest.NewServer(s.h.Handler)
	defer ts.Close()

	workspaceID := createWorkspaceViaHTTP(t, ts, wd)
	sessionID := createSessionViaHTTP(t, ts, workspaceID)

	status := "active"
	setReq := proto.ThreadGoalSetRequest{
		SessionID: sessionID,
		Objective: "port goal parity",
		Status:    &status,
	}
	setResp := postJSON[proto.ThreadGoalSetResponse](t, ts, "/v1/workspaces/"+workspaceID+"/thread/goal/set", setReq)
	require.NotNil(t, setResp.Goal)
	require.Equal(t, "port goal parity", setResp.Goal.Objective)
	require.Equal(t, "active", setResp.Goal.Status)

	getResp := getJSON[proto.ThreadGoalGetResponse](t, ts, "/v1/workspaces/"+workspaceID+"/thread/goal/get?session_id="+sessionID)
	require.NotNil(t, getResp.Goal)
	require.Equal(t, "port goal parity", getResp.Goal.Objective)
	require.Equal(t, "active", getResp.Goal.Status)

	clearResp := postJSON[proto.ThreadGoalClearResponse](t, ts, "/v1/workspaces/"+workspaceID+"/thread/goal/clear", proto.ThreadGoalClearRequest{
		SessionID: sessionID,
	})
	require.True(t, clearResp.Cleared)

	getAfterClear := getJSON[proto.ThreadGoalGetResponse](t, ts, "/v1/workspaces/"+workspaceID+"/thread/goal/get?session_id="+sessionID)
	require.Nil(t, getAfterClear.Goal)

	require.NoError(t, s.Shutdown(context.Background()))
}

func createWorkspaceViaHTTP(t *testing.T, ts *httptest.Server, wd string) string {
	t.Helper()
	req := proto.Workspace{
		Path:    wd,
		DataDir: filepath.Join(wd, ".crush"),
	}
	out := postJSON[proto.Workspace](t, ts, "/v1/workspaces", req)
	require.NotEmpty(t, out.ID)
	return out.ID
}

func createSessionViaHTTP(t *testing.T, ts *httptest.Server, workspaceID string) string {
	t.Helper()
	req := proto.Session{Title: "goal e2e"}
	out := postJSON[proto.Session](t, ts, "/v1/workspaces/"+workspaceID+"/sessions", req)
	require.NotEmpty(t, out.ID)
	return out.ID
}

func postJSON[T any](t *testing.T, ts *httptest.Server, path string, body any) T {
	t.Helper()
	var out T
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, ts.URL+path, bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

func getJSON[T any](t *testing.T, ts *httptest.Server, path string) T {
	t.Helper()
	var out T
	resp, err := http.Get(ts.URL + path) //nolint:gosec
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}
