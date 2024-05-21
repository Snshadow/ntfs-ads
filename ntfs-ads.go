//go:build windows
// +build windows

package ntfs_ads

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"

	"github.com/Snshadow/ntfs-ads/internal/w32api"
)

var (
	ErrNoADS       = errors.New("no alternate data stream found")
	ErrUnsupported = errors.New("file system does not support stream")
)

// parseStreamDataName parses ":streamname:$streamtype" format into name of stream.
// Returns stream name only if $streamtype is $DATA, otherwise returns empty string.
func parseStreamDataName(data w32api.WIN32_FIND_STREAM_DATA) string {
	wBuf := data.StreamName[:data.StreamSize]

	dataStr := windows.UTF16ToString(wBuf)

	fields := strings.Split(dataStr, ":")

	name, strmType := fields[1], fields[2]

	// not a data stream type
	if strmType != "$DATA" {
		return ""
	}

	return name
}

// GetFileADS finds names of alternate data streams from the named file.
func GetFileADS(path string) ([]string, error) {
	var adsNames []string

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	findStrm, data, err := w32api.FindFirstStream(absPath, w32api.FindStreamInfoStandard, 0)
	if err == windows.ERROR_HANDLE_EOF {
		return nil, ErrNoADS
	} else if err == windows.ERROR_INVALID_PARAMETER {
		return nil, ErrUnsupported
	} else if err != nil {
		return nil, err
	}

	if strmName := parseStreamDataName(data); strmName != "" {
		adsNames = append(adsNames, strmName)
	}

	for {
		data, findErr := w32api.FindNextStream(findStrm)
		if findErr == windows.ERROR_HANDLE_EOF {
			// no more streams
			break
		} else if findErr != nil {
			err = findErr
			goto EXIT
		}

		if strmName := parseStreamDataName(data); strmName != "" {
			adsNames = append(adsNames, strmName)
		}
	}

EXIT:
	closeErr := w32api.FindClose(findStrm)
	if closeErr != nil {
		return adsNames, fmt.Errorf("error: %v, FindClose err: %v", err, closeErr)
	}

	return adsNames, err
}
