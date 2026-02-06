package api

import (
	"fmt"
	"io"
	"net/url"
)

// FileService handles file/blob API calls
type FileService struct {
	client *Client
}

// FileEntry represents a file or directory entry from blob/recursive API
type FileEntry struct {
	FilePath  string  `json:"filePath"`
	Extension string  `json:"extension"`
	Size      int64   `json:"size"`
	LfsOid    *string `json:"lfsOid"`
	LockedBy  *string `json:"lockedBy"`
}

// Name returns the file name from the path
func (f *FileEntry) Name() string {
	if f.FilePath == "" {
		return ""
	}
	// Get the last component of the path
	for i := len(f.FilePath) - 1; i >= 0; i-- {
		if f.FilePath[i] == '/' {
			return f.FilePath[i+1:]
		}
	}
	return f.FilePath
}

// FileContent represents file content response
type FileContent struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Encoding string `json:"encoding"` // base64 or text
	Content  string `json:"content"`
}

// List returns files and directories at the given path
func (s *FileService) List(owner, project, ref, path string) ([]FileEntry, error) {
	// GitFlic API: GET /project/{owner}/{project}/blob/recursive?commitHash={ref}&directory={path}
	// Returns a direct array of file entries
	apiPath := fmt.Sprintf("/project/%s/%s/blob/recursive",
		url.PathEscape(owner),
		url.PathEscape(project))

	params := url.Values{}
	params.Set("commitHash", ref)
	if path != "" && path != "/" {
		params.Set("directory", path)
	}
	params.Set("depth", "1") // Only immediate children

	apiPath += "?" + params.Encode()

	// API returns a direct array, not _embedded wrapper
	var entries []FileEntry
	if err := s.client.Get(apiPath, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// Get returns the content of a file
func (s *FileService) Get(owner, project, ref, path string) (*FileContent, error) {
	// GitFlic API: GET /project/{owner}/{project}/blob/download returns raw file bytes
	// We use the download endpoint and read content as string
	body, err := s.Download(owner, project, ref, path)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return &FileContent{
		Path:    path,
		Content: string(data),
		Size:    int64(len(data)),
	}, nil
}

// Download downloads a file as raw bytes
func (s *FileService) Download(owner, project, ref, path string) (io.ReadCloser, error) {
	// GitFlic API: GET /project/{owner}/{project}/blob/download?commitHash={ref}&file={path}
	apiPath := fmt.Sprintf("/project/%s/%s/blob/download",
		url.PathEscape(owner),
		url.PathEscape(project))

	params := url.Values{}
	params.Set("commitHash", ref)
	params.Set("file", path) // API uses "file", not "fileName"
	apiPath += "?" + params.Encode()

	body, _, err := s.client.DownloadFile(apiPath)
	return body, err
}
