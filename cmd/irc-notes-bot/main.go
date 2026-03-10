package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/identw/irc-notes-bot/pkg/bot"
	"github.com/identw/irc-notes-bot/pkg/config"
	"github.com/identw/irc-notes-bot/pkg/db"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("irc-notes-bot %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("IRC Notes Bot %s starting...", version)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	log.Printf("Configuration loaded: server=%s:%d, channels=%v, TLS=%v",
		cfg.Server, cfg.Port, cfg.Channels, cfg.TLS)

	// Initialize notes store
	store, err := db.NewNoteStore(cfg.DBPath, cfg.MaxNotes)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer store.Close()
	log.Printf("Database initialized: %s (max notes=%d, max size=%d)",
		cfg.DBPath, cfg.MaxNotes, cfg.MaxNoteSize)

	// Create bot
	b, err := bot.New(cfg, store)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %s, shutting down...", sig)
		b.Client.Close()
		store.Close()
		os.Exit(0)
	}()

	// Start the bot
	if err := b.Run(); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
