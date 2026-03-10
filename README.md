# irc-notes-bot

IRC bot that stores per-channel notes in a SQLite database and sends them to users via private messages when they join a channel.

## Features

- Connects to IRC with or without TLS
- TLS supports both system trust store and a custom CA certificate
- Optional server password authentication
- Per-channel notes stored in SQLite (ring buffer — oldest notes are automatically removed when the limit is reached)
- Sends all channel notes to a user via PM on channel join
- Configurable limits for note count and note size

## Build

```bash
go build -o irc-notes-bot ./cmd/irc-notes-bot/
```

## Usage

```bash
./irc-notes-bot [-config <path>]
```

### Command-line flags

| Flag | Default | Description |
|---|---|---|
| `-config` | `config.yaml` | Path to the YAML configuration file |

## Configuration

The bot reads its configuration from a YAML file (`config.yaml` by default).

### Example

```yaml
server: "irc.example.com"
port: 6697
password: ""

nick: "notesbot"
user: "notesbot"
realname: "IRC Notes Bot"

tls: true
tls_ca: ""
tls_skip_verify: false

channels:
  - "#general"
  - "#dev"

db_path: "notes.db"
max_notes: 15
max_note_size: 4096
```

### Parameters

| Parameter | Type | Default | Required | Description |
|---|---|---|---|---|
| `server` | string | — | **yes** | IRC server hostname |
| `port` | int | `6667` | no | IRC server port |
| `password` | string | `""` | no | Server password (PASS) |
| `nick` | string | `"notesbot"` | no | Bot nickname |
| `user` | string | `"notesbot"` | no | Bot username (ident) |
| `realname` | string | `"IRC Notes Bot"` | no | Bot real name (GECOS) |
| `tls` | bool | `false` | no | Enable TLS connection |
| `tls_ca` | string | `""` | no | Path to a custom CA certificate file (PEM). If empty, the system trust store is used |
| `tls_skip_verify` | bool | `false` | no | Skip TLS certificate verification (not recommended for production) |
| `channels` | list | — | **yes** | List of channels the bot will join |
| `db_path` | string | `"notes.db"` | no | Path to the SQLite database file |
| `max_notes` | int | `15` | no | Maximum number of notes per channel (ring buffer) |
| `max_note_size` | int | `4096` | no | Maximum size of a single note in bytes |

## IRC Commands

All commands are issued in a channel where the bot is present:

| Command | Description |
|---|---|
| `!note add <text>` | Add a new note for the current channel |
| `!note list` | List all notes for the current channel |
| `!note help` | Show help message |

Quotes around the text are optional and will be stripped automatically:

```
!note add "remember to update the docs"
!note add remember to update the docs
```

Both forms are equivalent.

## Behavior

- When a user joins a channel, the bot sends all saved notes for that channel to the user via private message.
- Notes are stored per channel. Each channel has its own independent ring buffer.
- When the number of notes in a channel reaches `max_notes`, the oldest note is automatically deleted upon adding a new one.

## Project Structure

```
├── cmd/
│   └── irc-notes-bot/
│       └── main.go              # Entrypoint
├── pkg/
│   ├── bot/
│   │   └── bot.go               # IRC bot logic & event handlers
│   ├── config/
│   │   └── config.go            # YAML configuration loader
│   └── db/
│       └── db.go                # SQLite notes storage
├── config.yaml                  # Example configuration
├── go.mod
└── go.sum
```

## License

MIT
