package main

import (
	"log"
	"runtime"

	"qoder-sm"
)

func init() {
	// macOS需要运行在主线程
	runtime.LockOSThread()
}

func main() {
	log.SetFlags(log.Lshortfile)

	gui := qoder_sm.NewQoderGUI()
	gui.Run()
}
