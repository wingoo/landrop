package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/mdp/qrterminal/v3"

	"github.com/wingoo/landrop/internal/clip"
	"github.com/wingoo/landrop/internal/netutil"
	"github.com/wingoo/landrop/internal/server"
	"github.com/wingoo/landrop/internal/token"
)

const version = "v0.1.0"

func main() {
	port := flag.Int("port", 7777, "listen port")
	dir := flag.String("dir", defaultSaveDir(), "save directory")
	once := flag.Bool("once", false, "exit after first successful receive")
	textStdout := flag.Bool("text-stdout", false, "print received text to CLI and do not save text file")
	clipboard := flag.Bool("clipboard", false, "enable server-side clipboard copy (macOS/Windows)")
	noToken := flag.Bool("no-token", false, "disable token check (not recommended)")
	noQR := flag.Bool("no-qr", false, "disable printing QR code")
	flag.Parse()

	if err := os.MkdirAll(*dir, 0o755); err != nil {
		log.Fatalf("create dir: %v", err)
	}

	tok := ""
	if !*noToken {
		var err error
		tok, err = token.Generate(0)
		if err != nil {
			log.Fatalf("generate token: %v", err)
		}
	}

	if *clipboard && !clip.Supported() {
		log.Printf("warning: clipboard is not supported on %s, text will still be saved as file", runtime.GOOS)
	}

	lanIP := netutil.PrimaryIPv4()
	listenAddr := fmt.Sprintf("0.0.0.0:%d", *port)
	baseURL := fmt.Sprintf("http://%s:%d/", lanIP, *port)
	accessURL := baseURL
	if tok != "" {
		accessURL = fmt.Sprintf("%s?t=%s", baseURL, tok)
	}

	srv := server.New(server.Config{
		SaveDir:          *dir,
		Token:            tok,
		Once:             *once,
		TextToStdout:     *textStdout,
		ClipboardEnabled: *clipboard,
		ClipboardReady:   clip.Supported(),
		CopyToClipboard:  clip.CopyText,
		MaxBodyBytes:     200 * 1024 * 1024,
	})

	httpServer := &http.Server{
		Addr:              listenAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	srv.SetOnFirstSuccess(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	})

	printStartup(version, *dir, listenAddr, accessURL, tok)
	if !*noQR {
		fmt.Println("QR:")
		qrterminal.GenerateHalfBlock(accessURL, qrterminal.L, os.Stdout)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("received signal: %s", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}

	log.Println("landrop stopped")
}

func printStartup(version, dir, listenAddr, accessURL, token string) {
	fmt.Printf("landrop %s\n\n", version)
	fmt.Printf("Save dir:\n  %s\n\n", dir)
	fmt.Printf("Listening:\n  http://%s\n\n", listenAddr)
	fmt.Printf("Access:\n  %s\n\n", accessURL)
	fmt.Println("Endpoints:")
	if token == "" {
		fmt.Println("  POST /upload")
		fmt.Println("  POST /text")
		return
	}
	fmt.Printf("  POST /upload?t=%s\n", token)
	fmt.Printf("  POST /text?t=%s\n", token)
}

func defaultSaveDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, "Downloads", "landrop")
}
