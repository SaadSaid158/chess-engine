package main

import (
	"chessengine/engine"
	"chessengine/web"
	"flag"
	"fmt"
)

func main() {
	mode := flag.String("mode", "web", "Run mode: 'uci' for UCI protocol, 'web' for browser UI")
	port := flag.Int("port", 8080, "Port for web UI")
	flag.Parse()

	switch *mode {
	case "uci":
		uci := engine.NewUCI()
		uci.Loop()
	case "web":
		fmt.Println("Starting GoChess Web UI...")
		server := web.NewServer()
		server.Start(*port)
	default:
		fmt.Println("Unknown mode. Use 'uci' or 'web'.")
	}
}
