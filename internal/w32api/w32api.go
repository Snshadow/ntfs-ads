//go:build windows
// +build windows

package w32api

import (
	"bytes"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

//sys findFirstStream(fileName *uint16, infoLevel int32, findStreamData unsafe.Pointer, flags uint32) (hnd windows.Handle, err error) [failretval==windows.InvalidHandle] = kernel32.FindFirstStreamW
//sys findNextStream(findStream windows.Handle, findStreamData unsafe.Pointer) (err error) = kernel32.FindNextStreamW
//sys findClose(findFile windows.Handle) (err error) = kernel32.FindClose

func FindFirstStream(fileName string, infoLevel int32, flags uint32) (hnd windows.Handle, data WIN32_FIND_STREAM_DATA, err error) {
	wStr, err := windows.UTF16PtrFromString(fileName)
	if err != nil {
		return
	}

	hnd, err = findFirstStream(wStr, FindStreamInfoStandard, unsafe.Pointer(&data), flags) // flags should be 0

	// returned error
	// windows.ERROR_HANDLE_EOF if there is no stream
	// windows.ERROR_INVALID_PARAMETER for unsupported file system

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

/*
	typedef struct _FILE_RENAME_INFO {
	#if _WIN32_WINNT >= _WIN32_WINNT_WIN10_RS1
		__C89_NAMELESS union {
			BOOLEAN ReplaceIfExists;
			DWORD Flags;
		};
	#else
		BOOLEAN ReplaceIfExists;
	#endif
		HANDLE RootDirectory;
		DWORD FileNameLength;
		WCHAR FileName[1];
	} FILE_RENAME_INFO,*PFILE_RENAME_INFO;
*/

// NewFileRenameInfo returns FILE_RENAME_INFO from winbase.h as bytes buffer used for renaming ADS name.
func NewFileRenameInfo(newName string, replace bool) ([]byte, error) {
	if len(newName) == 0 {
		return nil, fmt.Errorf("new name is empty")
	}

	verInfo := windows.RtlGetVersion()

	var renameInfo bytes.Buffer
	var replaceIfExists uint32

	if replace {
		replaceIfExists = 1 // TRUE
	}

	if verInfo.MajorVersion < 10 || (verInfo.BuildNumber&0xffff) < 14393 { // under WIndows 10 Redstone 1
		renameInfo.WriteByte(byte(replaceIfExists)) // set ReplaceIfExists BOOLEAN(bool) to true
	} else {
		renameInfo.Write(unsafe.Slice((*byte)(unsafe.Pointer(&replaceIfExists)), unsafe.Sizeof(replaceIfExists))) // set ReplaceIfExists DWORD(uint32) to true
	}

	rootDirectory := windows.Handle(0)
	renameInfo.Write(unsafe.Slice((*byte)(unsafe.Pointer(&rootDirectory)), unsafe.Sizeof(rootDirectory))) // RootDirectory HANDLE(NULL)

	u16Name, err := windows.UTF16FromString(newName)
	if err != nil {
		return nil, err
	}

	var fileNameLength uint32 = uint32((len(u16Name) - 1) * 2) // length in bytes without NUL-terminaton

	renameInfo.Write(unsafe.Slice((*byte)(unsafe.Pointer(&fileNameLength)), unsafe.Sizeof(fileNameLength))) // FileNameLength DWORD(uint32)
	renameInfo.Write(unsafe.Slice((*byte)(unsafe.Pointer(&u16Name[0])), len(u16Name)*2)) // FileName []uint16(WCHAR[])

	return renameInfo.Bytes(), nil
}
