package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

func main() {
	path := "D:\\temp"
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return
	}
	handle, err := windows.CreateFile(
		p,
		syscall.FILE_LIST_DIRECTORY,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OVERLAPPED,
		0,
	)
	if err != nil {
		fmt.Println("CreateFile error:", err)
		return
	}
	defer windows.CloseHandle(handle)

	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		fmt.Println("CreateEvent error:", syscall.GetLastError())
		return
	}
	defer windows.CloseHandle(event)

	var overlapped windows.Overlapped
	overlapped.HEvent = event

	buffer := make([]byte, 4096)

	for {
		n := uint32(0)
		err := windows.ReadDirectoryChanges(
			handle,
			&buffer[0],
			uint32(len(buffer)),
			true,
			syscall.FILE_NOTIFY_CHANGE_FILE_NAME|
				syscall.FILE_NOTIFY_CHANGE_DIR_NAME|
				syscall.FILE_NOTIFY_CHANGE_ATTRIBUTES|
				syscall.FILE_NOTIFY_CHANGE_SIZE|
				syscall.FILE_NOTIFY_CHANGE_LAST_WRITE|
				syscall.FILE_NOTIFY_CHANGE_LAST_ACCESS|
				syscall.FILE_NOTIFY_CHANGE_CREATION,
			&n,
			&overlapped,
			0,
		)
		if err != nil && err != syscall.ERROR_IO_PENDING {
			fmt.Println("ReadDirectoryChangesW error:", err)
			return
		}

		syscall.WaitForSingleObject(syscall.Handle(event), syscall.INFINITE)

		var bytesTransferred uint32
		err = windows.GetOverlappedResult(handle, &overlapped, &bytesTransferred, true)
		if err != nil {
			fmt.Println("GetOverlappedResult error:", err)
			return
		}

		index := uint32(0)
		for {
			if index >= bytesTransferred {
				break
			}

			info := (*syscall.FileNotifyInformation)(unsafe.Pointer(&buffer[index]))
			filename := syscall.UTF16ToString((*[1 << 16]uint16)(unsafe.Pointer(&info.FileName))[:int(info.FileNameLength)/2])
			if filename == "" {
				index += info.NextEntryOffset
				continue
			}
			fmt.Println(filename, info.Action)
			if info.NextEntryOffset == 0 {
				break
			}
			index += info.NextEntryOffset
		}

	}
}
