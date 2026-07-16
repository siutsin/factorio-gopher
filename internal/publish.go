package gopher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	modPortalAPIKeyEnv = "FACTORIO_MOD_PORTAL_API_KEY" //nolint:gosec // environment variable name, not a credential
	modPortalInitURL   = "https://mods.factorio.com/api/v2/mods/releases/init_upload"
)

var modPortalHTTPClient = &http.Client{Timeout: 15 * time.Minute}

// Publish uploads a packaged mod release to the Factorio Mod Portal.
func Publish(args []string) (string, int, error) {
	return publish(args, os.Getenv, modPortalHTTPClient, modPortalInitURL)
}

func publish(
	args []string,
	getenv func(string) string,
	client *http.Client,
	initURL string,
) (string, int, error) {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	modDir := fs.String("mod", DefaultModDir(), "mod source directory containing info.json")
	archivePath := fs.String("archive", "", "mod zip (defaults to build/<name>_<version>.zip)")

	if err := fs.Parse(args); err != nil {
		return "", 2, err
	}
	if fs.NArg() != 0 {
		return "", 2, errors.New("usage: publish [-mod <dir>] [-archive <zip>]")
	}

	apiKey := strings.TrimSpace(getenv(modPortalAPIKeyEnv))
	if apiKey == "" {
		return "", 1, fmt.Errorf("%s is not set", modPortalAPIKeyEnv)
	}

	info, err := readModInfo(*modDir)
	if err != nil {
		return "", 1, err
	}
	if *archivePath == "" {
		*archivePath = filepath.Join(
			filepath.Dir(filepath.Clean(*modDir)),
			"build",
			fmt.Sprintf("%s_%s.zip", info.Name, info.Version),
		)
	}

	if err := publishRelease(
		context.Background(),
		client,
		initURL,
		apiKey,
		info.Name,
		*archivePath,
	); err != nil {
		return "", 1, fmt.Errorf("publish %s %s: %w", info.Name, info.Version, err)
	}

	return fmt.Sprintf("published %s %s", info.Name, info.Version), 0, nil
}

func publishRelease(
	ctx context.Context,
	client *http.Client,
	initURL, apiKey, modName, archivePath string,
) error {
	archive, err := os.ReadFile(archivePath) //nolint:gosec // maintainer-supplied release path
	if err != nil {
		return fmt.Errorf("read archive: %w", err)
	}

	initRequest, err := newMultipartRequest(ctx, initURL, func(writer *multipart.Writer) error {
		return writer.WriteField("mod", modName)
	})
	if err != nil {
		return fmt.Errorf("create init request: %w", err)
	}
	initRequest.Header.Set("Authorization", "Bearer "+apiKey)

	initResponse, err := client.Do(initRequest)
	if err != nil {
		return redactRequestURL("init upload", err)
	}
	var initResult struct {
		UploadURL string `json:"upload_url"`
		Error     string `json:"error"`
		Message   string `json:"message"`
	}
	if decodeErr := decodePortalResponse("init upload", initResponse, &initResult); decodeErr != nil {
		return decodeErr
	}
	if initResult.UploadURL == "" {
		return fmt.Errorf("init upload rejected: %s: %s", initResult.Error, initResult.Message)
	}

	uploadRequest, err := newMultipartRequest(ctx, initResult.UploadURL, func(writer *multipart.Writer) error {
		part, partErr := writer.CreateFormFile("file", filepath.Base(archivePath))
		if partErr == nil {
			_, partErr = part.Write(archive)
		}
		return partErr
	})
	if err != nil {
		return fmt.Errorf("create upload request: %w", err)
	}

	uploadResponse, err := client.Do(uploadRequest)
	if err != nil {
		return redactRequestURL("upload release", err)
	}
	var uploadResult struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := decodePortalResponse("upload release", uploadResponse, &uploadResult); err != nil {
		return err
	}
	if !uploadResult.Success {
		return fmt.Errorf("upload release rejected: %s: %s", uploadResult.Error, uploadResult.Message)
	}
	return nil
}

func redactRequestURL(label string, err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}
	return fmt.Errorf("%s: %w", label, err)
}

func newMultipartRequest(
	ctx context.Context,
	target string,
	write func(*multipart.Writer) error,
) (*http.Request, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := write(writer); err != nil {
		return nil, err
	}
	// bytes.Buffer writes cannot fail.
	_ = writer.Close() //nolint:errcheck // bytes.Buffer writes cannot fail

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target, &body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request, nil
}

func decodePortalResponse(label string, response *http.Response, target any) error {
	defer response.Body.Close() //nolint:errcheck // response body close errors do not affect decoded data
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("%s response: %w", label, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s: HTTP %s: %s", label, response.Status, strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode %s response: %w", label, err)
	}
	return nil
}
