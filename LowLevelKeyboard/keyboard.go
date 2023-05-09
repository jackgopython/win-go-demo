package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	WH_KEYBOARD_LL   = 13
	MAPVK_VK_TO_CHAR = 2
)

type KBDLLHOOKSTRUCT struct {
	vkCode      uint32
	scanCode    uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

var (
	mod                     = syscall.NewLazyDLL("user32.dll")
	procUnhookWindowsHookEx = mod.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = mod.NewProc("CallNextHookEx")
	procMapVirtualKeyEx     = mod.NewProc("MapVirtualKeyExW")
	hookHandle              uintptr
)

func LowLevelKeyboardProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 && wParam == 0x100 {
		kbdHook := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		charCode, _, _ := procMapVirtualKeyEx.Call(uintptr(kbdHook.vkCode), MAPVK_VK_TO_CHAR, 0)
		fmt.Printf("Keyboard event: vkCode=%d scanCode=%d flags=%d time=%d char='%c'\n",
			kbdHook.vkCode, kbdHook.scanCode, kbdHook.flags, kbdHook.time, rune(charCode))
	}
	r, _, _ := procCallNextHookEx.Call(hookHandle, uintptr(nCode), wParam, lParam)

	return r
}

func main() {
	user32 := syscall.NewLazyDLL("user32.dll")
	hook := user32.NewProc("SetWindowsHookExW")

	hookHandle, _, _ = hook.Call(
		WH_KEYBOARD_LL,
		syscall.NewCallback(LowLevelKeyboardProc),
		0,
		0,
	)
	if hookHandle == 0 {
		fmt.Println("Failed to set keyboard hook")
		return
	}

	var msg struct {
		hwnd   uintptr
		msg    uint32
		wParam uintptr
		lParam uintptr
		time   uint32
		pt     struct {
			x, y int32
		}
	}

	for GetMessage(&msg, 0, 0, 0) > 0 {
		TranslateMessage(&msg)
		DispatchMessage(&msg)
	}

	procUnhookWindowsHookEx.Call(hookHandle)
}

func GetMessage(msg *struct {
	hwnd   uintptr
	msg    uint32
	wParam uintptr
	lParam uintptr
	time   uint32
	pt     struct {
		x, y int32
	}
}, hWnd uintptr, wMsgFilterMin, wMsgFilterMax uint32) int32 {
	ret, _, _ := syscall.Syscall6(getMessage.Addr(), 4,
		uintptr(unsafe.Pointer(msg)),
		hWnd,
		uintptr(wMsgFilterMin),
		uintptr(wMsgFilterMax),
		0,
		0)

	return int32(ret)
}

func TranslateMessage(msg *struct {
	hwnd   uintptr
	msg    uint32
	wParam uintptr
	lParam uintptr
	time   uint32
	pt     struct {
		x, y int32
	}
}) bool {
	ret, _, _ := syscall.Syscall(translateMessage.Addr(), 1,
		uintptr(unsafe.Pointer(msg)),
		0,
		0)

	return ret != 0
}

func DispatchMessage(msg *struct {
	hwnd   uintptr
	msg    uint32
	wParam uintptr
	lParam uintptr
	time   uint32
	pt     struct {
		x, y int32
	}
}) uintptr {
	ret, _, _ := syscall.Syscall(dispatchMessage.Addr(), 1,
		uintptr(unsafe.Pointer(msg)),
		0,
		0)

	return ret
}

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	getMessage       = user32.NewProc("GetMessageW")
	translateMessage = user32.NewProc("TranslateMessage")
	dispatchMessage  = user32.NewProc("DispatchMessageW")
)
