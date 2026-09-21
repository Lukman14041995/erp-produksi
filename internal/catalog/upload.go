package catalog

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/ranji/clothing-erp/internal/apperr"
)

var allowedImageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// SaveCatalogImage writes an uploaded photo (fabric sample, ink print
// sample, or a garment variant's icon) to <uploadDir>/<subdir> and returns
// the URL path it's served under (see e.Static("/uploads", ...) in
// cmd/api/main.go and the matching nginx /uploads/ proxy). The stored
// filename is always server-generated (never the client's original name),
// so there's no path-traversal or extension-spoofing surface from the
// upload itself.
func SaveCatalogImage(fh *multipart.FileHeader, subdir, uploadDir string) (string, error) {
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedImageExt[ext] {
		return "", apperr.Validation("image must be one of: jpg, jpeg, png, webp")
	}

	src, err := fh.Open()
	if err != nil {
		return "", apperr.Internal("open uploaded file", err)
	}
	defer src.Close()

	destDir := filepath.Join(uploadDir, subdir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", apperr.Internal("create upload directory", err)
	}

	filename := uuid.NewString() + ext
	destPath := filepath.Join(destDir, filename)

	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", apperr.Internal("create destination file", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", apperr.Internal("write uploaded file", err)
	}

	return fmt.Sprintf("/uploads/%s/%s", subdir, filename), nil
}
