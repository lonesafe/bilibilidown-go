package main

import (
	"fmt"
	"log"
	"net/http"

	"bilibilidown-go/internal/config"
	"bilibilidown-go/internal/server"
)

func main() {
	cfg := config.Get()
	srv := server.NewServer()
	handler := srv.RegisterRoutes()

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	fmt.Printf("========================================\n")
	fmt.Printf("  BilibiliDown Go Server\n")
	fmt.Printf("  监听地址: http://localhost%s\n", addr)
	fmt.Printf("  下载路径: %s\n", cfg.SavePath)
	fmt.Printf("========================================\n")

	log.Fatal(http.ListenAndServe(addr, handler))
}
