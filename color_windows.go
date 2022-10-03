package zlog

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

// put powershell into ANSI color mode
func init() {
	envTerm := os.Getenv("TERM")
	// xterm and friends, nothing to do
	if strings.Contains(envTerm, "term") {
		return
	}

	SupportColors = false
	outHandle, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		fmt.Println(53, err)
		return
	}

	// load related windows dll
	// isMSys = utils.IsMSys()
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	//procGetConsoleMode := kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode := kernel32.NewProc("SetConsoleMode")

	var mode uint32
	if err = syscall.GetConsoleMode(outHandle, &mode); err != nil {
		fmt.Println(92, err)
		return
	}

	const EnableVirtualTerminalProcessingMode uint32 = 0x4
	mode |= EnableVirtualTerminalProcessingMode

	procSetConsoleMode.Call(uintptr(outHandle), uintptr(mode))
	SupportColors = true
}
