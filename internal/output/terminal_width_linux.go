package output

import (
	"os"
	"syscall"
	"unsafe"
)

type windowSize struct {
	rows    uint16
	columns uint16
	xPixels uint16
	yPixels uint16
}

func terminalWidthFromFile(file *os.File) (int, bool) {
	if file == nil {
		return 0, false
	}
	var size windowSize
	operation := uintptr(syscall.SYS_IOCTL)
	fileDescriptor := file.Fd()
	request := uintptr(syscall.TIOCGWINSZ)
	destination := uintptr(unsafe.Pointer(&size))
	_, _, errno := syscall.Syscall(operation, fileDescriptor, request, destination)
	if errno != 0 {
		return 0, false
	}
	if size.columns != 0 {
		newSize := int(size.columns)
		return newSize, true
	}

	return 0, false
}
