//go:build windows
// +build windows

package w32api

import (
	"golang.org/x/sys/windows"
)

const (
	FindStreamInfoStandard = 0
)

type WIN32_FIND_STREAM_DATA struct {
	StreamSize int64
	StreamName [windows.MAX_PATH + 36]uint16 // ":streamname:$streamtype", possible $streamtype: $DATA, $INDEX_ALLOCATION, $BITMAP
}
