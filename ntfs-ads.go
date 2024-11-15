//go:build windows
// +build windows

package ntfs_ads

import (
	"errors"
	"fmt"
	"os"
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
	dataStr := windows.UTF16ToString(data.StreamName[:])

	fields := strings.Split(dataStr, ":")

	name, strmType := fields[1], fields[2]

	// not a data stream type
	if strmType != "$DATA" {
		return ""
	}

	return name
}

// OpenFileADS opens data stream of the name from the given file with specified flag(see os.OpenFile() for details), should be closed with (*os.File).Close() after use.
func OpenFileADS(path string, name string, openFlag int) (*os.File, error) {
	path = path + ":" + name

	u16Path, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	var access, mode, createmode uint32

	switch openFlag & (os.O_RDONLY | os.O_WRONLY | os.O_RDWR) {
	case os.O_RDONLY:
		access = windows.FILE_READ_DATA | windows.SYNCHRONIZE
		mode = windows.FILE_SHARE_READ
	case os.O_WRONLY:
		access = windows.FILE_WRITE_DATA | windows.SYNCHRONIZE
		mode = windows.FILE_SHARE_WRITE
	case os.O_RDWR:
		access = windows.FILE_READ_DATA | windows.FILE_WRITE_DATA | windows.SYNCHRONIZE
		mode = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE
	}

	switch openFlag & (os.O_CREATE | os.O_TRUNC | os.O_EXCL) {
	case os.O_CREATE | os.O_EXCL:
		createmode = windows.CREATE_NEW
	case os.O_CREATE | os.O_TRUNC:
		createmode = windows.CREATE_ALWAYS
	case os.O_CREATE:
		createmode = windows.OPEN_ALWAYS
	case os.O_TRUNC:
		createmode = windows.TRUNCATE_EXISTING
	default:
		createmode = windows.OPEN_EXISTING
	}

	if openFlag&os.O_APPEND != 0 {
		access &^= windows.FILE_WRITE_DATA
		access |= windows.FILE_APPEND_DATA
	}

	hnd, err := windows.CreateFile(
		u16Path,
		access,
		mode,
		nil,
		createmode,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(hnd), path), nil
}

// GetFileADS finds names of alternate data streams from the named file.
func GetFileADS(path string) (map[string]int64, error) {
	streamInfoMap := make(map[string]int64)

	var err error
	var absPath string

	if strings.HasPrefix(path, "\\??\\") {
		// has NT Namespace prefix
		absPath = path
	} else {
		absPath, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
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
		streamInfoMap[strmName] = data.StreamSize
	}

	for {
		data, findErr := w32api.FindNextStream(findStrm)
		if findErr == windows.ERROR_HANDLE_EOF {
			// no more stream
			break
		} else if findErr != nil {
			err = findErr
			goto EXIT
		}

		if strmName := parseStreamDataName(data); strmName != "" {
			streamInfoMap[strmName] = data.StreamSize
		}
	}

EXIT:
	closeErr := w32api.FindClose(findStrm)
	if closeErr != nil {
		return streamInfoMap, fmt.Errorf("error: %v, FindClose err: %v", err, closeErr)
	}

	return streamInfoMap, err
}
