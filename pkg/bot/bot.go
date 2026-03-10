package bot

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lrstanley/girc"

	"github.com/identw/irc-notes-bot/pkg/config"
	"github.com/identw/irc-notes-bot/pkg/db"
)

// Bot represents an IRC bot with notes support.
type Bot struct {
	cfg    *config.Config
	Client *girc.Client
	store  *db.NoteStore
}

// New creates a new IRC bot.
func New(cfg *config.Config, store *db.NoteStore) (*Bot, error) {
	b := &Bot{
		cfg:   cfg,
		store: store,
	}

	ircCfg := girc.Config{
		Server: cfg.Server,
		Port:   cfg.Port,
		Nick:   cfg.Nick,
		User:   cfg.User,
		Name:   cfg.RealName,
	}

	if cfg.Password != "" {
		ircCfg.ServerPass = cfg.Password
	}

	if cfg.TLS {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: cfg.TLSSkipVerify,
		}

		// If CA path is specified — load it, otherwise use system trust store
		if cfg.TLSCA != "" {
			caCert, err := os.ReadFile(cfg.TLSCA)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA file %s: %w", cfg.TLSCA, err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to add CA certificate from %s", cfg.TLSCA)
			}
			tlsCfg.RootCAs = caCertPool
		}

		ircCfg.TLSConfig = tlsCfg
		ircCfg.SSL = true
	}

	client := girc.New(ircCfg)
	b.Client = client

	// Handler: join channels on connect
	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		log.Println("Connected to server, joining channels...")
		for _, ch := range cfg.Channels {
			log.Printf("Joining channel %s", ch)
			c.Cmd.Join(ch)
		}
	})

	// Handler: when someone joins a channel — send them notes via PM
	client.Handlers.Add(girc.JOIN, func(c *girc.Client, e girc.Event) {
		b.handleJoin(c, e)
	})

	// Handler: channel commands
	client.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		b.handlePrivmsg(c, e)
	})

	// Handler: disconnection
	client.Handlers.Add(girc.DISCONNECTED, func(c *girc.Client, e girc.Event) {
		log.Println("Disconnected from server")
	})

	return b, nil
}

// Run starts the bot (blocking call).
func (b *Bot) Run() error {
	addr := fmt.Sprintf("%s:%d", b.cfg.Server, b.cfg.Port)
	log.Printf("Connecting to %s (TLS: %v)...", addr, b.cfg.TLS)
	return b.Client.Connect()
}

// handleJoin handles the JOIN event — sends notes to the user.
func (b *Bot) handleJoin(c *girc.Client, e girc.Event) {
	// Ignore our own JOIN
	if e.Source.Name == c.GetNick() {
		return
	}

	nick := e.Source.Name
	channel := e.Params[0]

	notes, err := b.store.ListNotes(channel)
	if err != nil {
		log.Printf("Error retrieving notes for %s: %v", channel, err)
		return
	}

	if len(notes) == 0 {
		return
	}

	// Send notes via private messages
	c.Cmd.Message(nick, fmt.Sprintf("📌 Channel %s — saved notes:", channel))
	for i, note := range notes {
		msg := fmt.Sprintf("  %d. [%s] (%s): %s",
			i+1,
			note.CreatedAt.Format("2006-01-02 15:04"),
			note.Author,
			note.Text,
		)
		c.Cmd.Message(nick, msg)
	}
	c.Cmd.Message(nick, "---")
	c.Cmd.Message(nick, "Commands: !note add <text> | !note list | !note help")
}

// handlePrivmsg handles messages in channels.
func (b *Bot) handlePrivmsg(c *girc.Client, e girc.Event) {
	// Only process messages in channels (starting with # or &)
	if len(e.Params) == 0 {
		return
	}
	target := e.Params[0]
	if !strings.HasPrefix(target, "#") && !strings.HasPrefix(target, "&") {
		return
	}

	channel := target
	text := e.Last()
	nick := e.Source.Name

	if !strings.HasPrefix(text, "!note ") && text != "!note" {
		return
	}

	// Parse command
	parts := strings.SplitN(strings.TrimPrefix(text, "!note"), " ", 2)
	cmd := strings.TrimSpace(parts[0])
	var arg string
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}

	switch cmd {
	case "add":
		b.cmdAdd(c, channel, nick, arg)
	case "list":
		b.cmdList(c, channel)
	case "help", "":
		b.cmdHelp(c, channel)
	default:
		c.Cmd.Message(channel, fmt.Sprintf("%s: unknown command. Use: !note help", nick))
	}
}

// cmdAdd adds a note.
func (b *Bot) cmdAdd(c *girc.Client, channel, nick, text string) {
	if text == "" {
		c.Cmd.Message(channel, fmt.Sprintf("%s: please provide text: !note add <text>", nick))
		return
	}

	// Strip quotes if present
	text = strings.Trim(text, "\"'")

	if len([]byte(text)) > b.cfg.MaxNoteSize {
		c.Cmd.Message(channel, fmt.Sprintf("%s: text is too long (max %d bytes)", nick, b.cfg.MaxNoteSize))
		return
	}

	if err := b.store.AddNote(channel, nick, text); err != nil {
		log.Printf("Error adding note: %v", err)
		c.Cmd.Message(channel, fmt.Sprintf("%s: error saving note", nick))
		return
	}

	count, _ := b.store.CountNotes(channel)
	c.Cmd.Message(channel, fmt.Sprintf("%s: ✅ note saved (total: %d/%d)", nick, count, b.cfg.MaxNotes))
}

// cmdList displays the list of notes in the channel.
func (b *Bot) cmdList(c *girc.Client, channel string) {
	notes, err := b.store.ListNotes(channel)
	if err != nil {
		log.Printf("Error retrieving notes: %v", err)
		c.Cmd.Message(channel, "Error retrieving notes list")
		return
	}

	if len(notes) == 0 {
		c.Cmd.Message(channel, "📭 No saved notes for this channel")
		return
	}

	c.Cmd.Message(channel, fmt.Sprintf("📌 Notes for %s (%d/%d):", channel, len(notes), b.cfg.MaxNotes))
	for i, note := range notes {
		msg := fmt.Sprintf("  %d. [%s] (%s): %s",
			i+1,
			note.CreatedAt.Format("2006-01-02 15:04"),
			note.Author,
			note.Text,
		)
		c.Cmd.Message(channel, msg)
	}
}

// cmdHelp displays command help.
func (b *Bot) cmdHelp(c *girc.Client, channel string) {
	c.Cmd.Message(channel, "📖 IRC Notes Bot — command help:")
	c.Cmd.Message(channel, "  !note add <text>   — add a note (max "+fmt.Sprintf("%d", b.cfg.MaxNoteSize)+" bytes)")
	c.Cmd.Message(channel, "  !note list         — show all notes for the channel")
	c.Cmd.Message(channel, "  !note help         — show this help")
	c.Cmd.Message(channel, fmt.Sprintf("  Limit: %d notes per channel (ring buffer)", b.cfg.MaxNotes))
	c.Cmd.Message(channel, "  On channel join you will receive all notes via PM")
}
