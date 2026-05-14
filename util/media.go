package util

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// MediaRoot returns the absolute path to the media storage root inside the repo
func MediaRoot() string {
	// Canonical uploads location at repo root
	mediaPath := viper.GetString("MEDIA_ROOT")
	if mediaPath == "" {
		return MustGitPath("uploads")
	}
	return mediaPath
}

// EnsureDir ensures the directory exists
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// PatientPhotoPath returns the absolute file path for a patient's photo.
// Stored as uploads/patient/<id> (no extension); MIME is detected at read time.
func PatientPhotoPath(id int32) string {
	return filepath.Join(MediaRoot(), "patient", fmt.Sprintf("%d", id))
}

// MaxPhotoSize is the maximum allowed size of an uploaded photo (5 MiB)
const MaxPhotoSize = 5 << 20

// ValidateImageBytes checks basic constraints and returns the detected MIME type
func ValidateImageBytes(data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty file")
	}
	if len(data) > MaxPhotoSize {
		return "", errors.New("file too large")
	}
	probe := len(data)
	if probe > 512 {
		probe = 512
	}
	mime := http.DetectContentType(data[:probe])
	switch mime {
	case "image/jpeg", "image/png", "image/webp", "image/gif", "image/bmp", "application/octet-stream":
		return mime, nil
	default:
		return "", errors.New("unsupported image type")
	}
}

// SavePatientPhoto writes the bytes to the deterministic path for the patient
func SavePatientPhoto(id int32, data []byte) error {
	abs := PatientPhotoPath(id)
	dir := filepath.Dir(abs)
	if err := EnsureDir(dir); err != nil {
		return err
	}
	tmp := abs + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, abs); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// DeletePatientPhotoIfExists removes the image file if it exists
func DeletePatientPhotoIfExists(id int32) error {
	abs := PatientPhotoPath(id)
	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Remove(abs)
}
