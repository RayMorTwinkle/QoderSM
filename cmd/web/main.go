package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"time"

	"qoder-sm/pkg/qoder"
)

var port = flag.Int("p", 8866, "Web服务器端口")

func main() {
	flag.Parse()

	server := qoder.NewWebServer(*port)

	// 在goroutine中启动服务器
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 打开浏览器
	url := fmt.Sprintf("http://localhost:%d", *port)
	openBrowser(url)

	// 保持程序运行
	select {}
}

func openBrowser(url string) {
	exec.Command("open", url).Start()
}
