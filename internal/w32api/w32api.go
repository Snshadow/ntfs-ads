//go:build windows
// +build windows

package w32api

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

//sys findFirstStream(fileName *uint16, infoLevel int32, findStreamData unsafe.Pointer, flags uint32) (hnd windows.Handle) = kernel32.FindFirstStreamW
//sys findNextStream(findStream windows.Handle, findStreamData unsafe.Pointer) (err error) = kernel32.FindNextStreamW
//sys findClose(findFile windows.Handle) (err error) = kernel32.FindClose

func FindFirstStream(fileName string, infoLevel int32, flags uint32) (hnd windows.Handle, data WIN32_FIND_STREAM_DATA, err error) {
	wStr, err := windows.UTF16PtrFromString(fileName)
	if err != nil {
		return
	}

	ret := findFirstStream(wStr, FindStreamInfoStandard, unsafe.Pointer(&data), flags) // flags should be 0
	if ret == windows.InvalidHandle {
		err = windows.GetLastError()
		return
	}

	// returns
	// windows.ERROR_HANDLE_EOF if there is no stream
	// windows.ERROR_INVALID_PARAMETER for unsupported file system

	hnd = ret

	return
}


func FindNextStream(findStream windows.Handle) (data WIN32_FIND_STREAM_DATA, err error) {
	err = findNextStream(findStream, unsafe.Pointer(&data))

	// returns windows.ERROR_HANDLE_EOF if there is no more stream

	return
}

func FindClose(findFile windows.Handle) (err error) {
	err = findClose(findFile)

	return
}
