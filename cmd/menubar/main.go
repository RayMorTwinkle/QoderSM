package main

import (
	"log"
	"runtime"

	"qoder-sm"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	log.SetFlags(log.Lshortfile)

	menubar := qoder_sm.NewQoderMenuBar()
	menubar.Run()
}
