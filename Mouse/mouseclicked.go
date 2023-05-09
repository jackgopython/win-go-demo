package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	WH_MOUSE_LL    = 14
	WM_LBUTTONDOWN = 0x0201
)

type POINT struct {
	X, Y int32
}

type MOUSEHOOKSTRUCT struct {
	Pt          POINT
	Hwnd        uintptr
	WHitTest    uint32
	DwExtraInfo uintptr
}

var (
	user32              = syscall.MustLoadDLL("user32.dll")
	SetWindowsHookEx    = user32.MustFindProc("SetWindowsHookExW")
	CallNextHookEx      = user32.MustFindProc("CallNextHookEx")
	UnhookWindowsHookEx = user32.MustFindProc("UnhookWindowsHookEx")
	GetMessage          = user32.MustFindProc("GetMessageW")
)

func main() {
	hook, _, _ := SetWindowsHookEx.Call(
		WH_MOUSE_LL,
		syscall.NewCallback(MouseProc),
		0,
		0)
	if hook == 0 {
		fmt.Println("Failed to set hook")
		return
	}
	fmt.Println("Hook set")
	defer UnhookWindowsHookEx.Call(hook)

	for {
		ret, _, _ := GetMessage.Call(0, 0, 0, 0)
		if ret == 0 {
			break
		}
		CallNextHookEx.Call(hook, 0, 0, 0)
	}
}

func MouseProc(nCode uintptr, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 && wParam == WM_LBUTTONDOWN {
		mouseHookStruct := (*MOUSEHOOKSTRUCT)(unsafe.Pointer(lParam))
		fmt.Printf("Mouse clicked at (%d, %d)\n", mouseHookStruct.Pt.X, mouseHookStruct.Pt.Y)
	}
	r, _, _ := CallNextHookEx.Call(0, nCode, wParam, lParam)

	return r

}
