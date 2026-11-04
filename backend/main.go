package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"flclashweb/backend/internal/api"
	"flclashweb/backend/internal/hub"
	"flclashweb/backend/internal/rpc"
	"flclashweb/backend/internal/store"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "HTTP listen address")
	home := flag.String("home", "", "data directory (default ~/.flclash-web)")
	flag.Parse()

	homeDir := *home
	if homeDir == "" {
		userConfig, err := os.UserConfigDir()
		if err != nil {
			userConfig = "."
		}
		homeDir = filepath.Join(userConfig, "flclash-web")
	}
	if err := os.MkdirAll(filepath.Join(homeDir, "profiles"), 0o755); err != nil {
		log.Fatalf("create home: %v", err)
	}

	db, err := store.Open(filepath.Join(homeDir, "flclash.db"))
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer db.Close()

	application := hub.New(homeDir, homeDir)
	application.SetStore(db)
	if err := application.LoadState(); err != nil {
		log.Printf("load state: %v", err)
	}

	client, err := rpc.Start(homeDir)
	if err != nil {
		log.Fatalf("start core: %v", err)
	}
	client.OnDisconnect = func() {
		application.EmitCoreDisconnected()
	}
	application.SetClient(client)
	application.StartEventPump()

	if err := application.InitCore(); err != nil {
		log.Printf("core init: %v", err)
	}
	application.StartTrafficPoller()
	application.StartAutoUpdater()

	// apply the selected profile in the background so the UI has proxies
	// right after boot, and auto-run the listener when configured.
	go func() {
		time.Sleep(100 * time.Millisecond)
		if err := application.ApplyProfile(true); err != nil {
			log.Printf("apply profile: %v", err)
		}
		if cfg := application.Config(); cfg.AppSetting.AutoRun {
			if err := application.StartListener(); err != nil {
				log.Printf("auto run: %v", err)
			}
		}
	}()

	static := api.StaticHandler()
	server := api.NewServer(application, client, static)

	fmt.Printf("FlClash Web listening on http://%s (data: %s)\n", *addr, homeDir)
	if err := http.ListenAndServe(*addr, server); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
