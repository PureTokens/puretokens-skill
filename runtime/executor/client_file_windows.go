package main

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

func clientPathSafe(_ string, info os.FileInfo) bool {
	attributes, ok := info.Sys().(*syscall.Win32FileAttributeData)
	return ok && attributes.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT == 0
}

func singleLinkFile(path string, info os.FileInfo) bool {
	if !info.Mode().IsRegular() || !clientPathSafe(path, info) {
		return false
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	handle, err := windows.CreateFile(name, windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)
	var details windows.ByHandleFileInformation
	return windows.GetFileInformationByHandle(handle, &details) == nil &&
		details.NumberOfLinks == 1 && details.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT == 0
}
