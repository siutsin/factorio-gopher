package gopher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishCLI(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "mod")
	buildDir := filepath.Join(root, "build")
	writeModInfo(t, modDir, "0.0.3")
	require.NoError(t, os.MkdirAll(buildDir, 0o750))
	archive := []byte("release zip")
	require.NoError(t, os.WriteFile(filepath.Join(buildDir, "gopher_0.0.3.zip"), archive, 0o600))

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		reader, err := r.MultipartReader()
		if !assert.NoError(t, err) {
			return
		}
		part, err := reader.NextPart()
		if !assert.NoError(t, err) {
			return
		}
		switch r.URL.Path {
		case "/init":
			assert.Equal(t, "Bearer secret", r.Header.Get("Authorization"))
			assert.Equal(t, "mod", part.FormName())
			got, err := io.ReadAll(part)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, "gopher", string(got))
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]string{"upload_url": server.URL + "/upload"}))
		case "/upload":
			assert.Equal(t, "file", part.FormName())
			got, err := io.ReadAll(part)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, "gopher_0.0.3.zip", part.FileName())
			assert.Equal(t, archive, got)
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]bool{"success": true}))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	message, exitCode, err := publish(
		[]string{"-mod", modDir},
		func(string) string { return "secret" },
		server.Client(),
		server.URL+"/init",
	)

	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "published gopher 0.0.3", message)
}

func TestPublishWrapper(t *testing.T) {
	t.Setenv(modPortalAPIKeyEnv, "")
	_, exitCode, err := Publish(nil)
	assert.Equal(t, 1, exitCode)
	assert.ErrorContains(t, err, modPortalAPIKeyEnv)
}

func TestPublishErrors(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "mod")
	writeModInfo(t, modDir, "0.0.3")

	cases := []struct {
		name      string
		args      []string
		getenv    func(string) string
		wantExit  int
		wantError string
	}{
		{
			name:      "flag parse error",
			args:      []string{"-unknown"},
			getenv:    func(string) string { return "secret" },
			wantExit:  2,
			wantError: "flag provided but not defined",
		},
		{
			name:      "unexpected positional argument",
			args:      []string{"extra"},
			getenv:    func(string) string { return "secret" },
			wantExit:  2,
			wantError: "usage: publish",
		},
		{
			name:      "missing API key",
			args:      []string{"-mod", modDir},
			getenv:    func(string) string { return "  " },
			wantExit:  1,
			wantError: modPortalAPIKeyEnv,
		},
		{
			name:      "missing info",
			args:      []string{"-mod", filepath.Join(root, "missing")},
			getenv:    func(string) string { return "secret" },
			wantExit:  1,
			wantError: "info.json",
		},
		{
			name:      "missing archive",
			args:      []string{"-mod", modDir, "-archive", filepath.Join(root, "missing.zip")},
			getenv:    func(string) string { return "secret" },
			wantExit:  1,
			wantError: "read archive",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, exitCode, err := publish(tc.args, tc.getenv, http.DefaultClient, modPortalInitURL)
			assert.Equal(t, tc.wantExit, exitCode)
			assert.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestPublishReleaseErrors(t *testing.T) { //nolint:gocognit // table covers distinct protocol failures
	archivePath := filepath.Join(t.TempDir(), "gopher.zip")
	require.NoError(t, os.WriteFile(archivePath, []byte("zip"), 0o600))

	t.Run("missing archive", func(t *testing.T) {
		err := publishRelease(
			context.Background(),
			http.DefaultClient,
			modPortalInitURL,
			"secret",
			"gopher",
			filepath.Join(t.TempDir(), "missing.zip"),
		)
		assert.ErrorContains(t, err, "read archive")
	})

	t.Run("invalid init URL", func(t *testing.T) {
		err := publishRelease(context.Background(), http.DefaultClient, "://", "secret", "gopher", archivePath)
		assert.ErrorContains(t, err, "create init request")
	})

	t.Run("init transport error", func(t *testing.T) {
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errBoom
		})}
		err := publishRelease(context.Background(), client, "https://example.invalid", "secret", "gopher", archivePath)
		assert.ErrorContains(t, err, "init upload")
	})

	cases := []struct {
		name       string
		initStatus int
		initBody   string
		uploadCode int
		uploadBody string
		want       string
	}{
		{name: "init HTTP error", initStatus: http.StatusForbidden, initBody: `{"error":"Forbidden"}`, want: "HTTP 403"},
		{name: "invalid init JSON", initBody: `{`, want: "decode init upload"},
		{name: "init rejected", initBody: `{"error":"Forbidden","message":"denied"}`, want: "init upload rejected"},
		{name: "upload HTTP error", initBody: `upload`, uploadCode: http.StatusBadRequest, uploadBody: `{"error":"InvalidModUpload"}`, want: "HTTP 400"},
		{name: "invalid upload JSON", initBody: `upload`, uploadBody: `{`, want: "decode upload release"},
		{name: "upload rejected", initBody: `upload`, uploadBody: `{"error":"InvalidModRelease","message":"duplicate"}`, want: "upload release rejected"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/init" {
					if tc.initStatus != 0 {
						w.WriteHeader(tc.initStatus)
					}
					if tc.initBody == "upload" {
						writeTestResponse(t, w, fmt.Sprintf(`{"upload_url":%q}`, server.URL+"/upload"))
						return
					}
					writeTestResponse(t, w, tc.initBody)
					return
				}
				if tc.uploadCode != 0 {
					w.WriteHeader(tc.uploadCode)
				}
				writeTestResponse(t, w, tc.uploadBody)
			}))
			t.Cleanup(server.Close)

			err := publishRelease(
				context.Background(),
				server.Client(),
				server.URL+"/init",
				"secret",
				"gopher",
				archivePath,
			)
			assert.ErrorContains(t, err, tc.want)
		})
	}

	t.Run("invalid upload URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeTestResponse(t, w, `{"upload_url":"://"}`)
		}))
		t.Cleanup(server.Close)
		err := publishRelease(context.Background(), server.Client(), server.URL, "secret", "gopher", archivePath)
		assert.ErrorContains(t, err, "create upload request")
	})

	t.Run("upload transport error", func(t *testing.T) {
		const uploadToken = "secret-upload-token"
		calls := 0
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return jsonResponse(
					http.StatusOK,
					`{"upload_url":"https://example.invalid/upload?token=`+uploadToken+`"}`,
				), nil
			}
			return nil, errBoom
		})}
		err := publishRelease(context.Background(), client, "https://example.invalid/init", "secret", "gopher", archivePath)
		require.ErrorContains(t, err, "upload release")
		assert.NotContains(t, err.Error(), uploadToken)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestMultipartAndResponseErrors(t *testing.T) {
	_, err := newMultipartRequest(
		context.Background(),
		"https://example.invalid",
		func(*multipart.Writer) error { return errBoom },
	)
	require.ErrorIs(t, err, errBoom)

	response := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(errorReader{}),
	}
	err = decodePortalResponse("test", response, &struct{}{})
	require.ErrorIs(t, err, errBoom)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func writeTestResponse(t *testing.T, writer io.Writer, body string) {
	t.Helper()
	_, err := io.WriteString(writer, body)
	assert.NoError(t, err)
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errBoom }
