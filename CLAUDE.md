# fastmail-cli

A command-line interface for Fastmail using the JMAP protocol, built with Go and Cobra.

## Project Structure

```
fastmail-cli/
├── cmd/fm/           # Main entry point
│   └── main.go
├── internal/
│   ├── cli/          # Cobra commands
│   │   ├── root.go
│   │   ├── inbox.go
│   │   ├── read.go
│   │   ├── search.go
│   │   ├── archive.go
│   │   ├── draft.go
│   │   ├── folders.go
│   │   └── ...
│   └── jmap/         # JMAP client wrapper
│       ├── client.go
│       ├── email.go
│       ├── mailbox.go
│       └── identity.go
├── node/             # Original TypeScript implementation (reference)
└── go.mod
```

## Dependencies

- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework with built-in shell completion
- [go-jmap](https://git.sr.ht/~rockorager/go-jmap) - JMAP client library

## Commands to Implement

Port from the TypeScript implementation in `node/src/`:

| Command | Description | Priority |
|---------|-------------|----------|
| `fm inbox` | List recent inbox emails | High |
| `fm read <id>` | Read a specific email | High |
| `fm search <query>` | Search emails | High |
| `fm thread <id>` | View email thread | Medium |
| `fm archive <id>` | Archive an email | High |
| `fm bulk-archive <query>` | Archive matching emails | Medium |
| `fm move <id> <folder>` | Move email to folder | Medium |
| `fm folders` | List mailboxes | High |
| `fm folder-create <name>` | Create folder | Low |
| `fm folder-rename <id> <name>` | Rename folder | Low |
| `fm draft` | Create draft | Medium |
| `fm reply <id>` | Create reply draft | Medium |
| `fm forward <id>` | Create forward draft | Medium |
| `fm attachment <id> <part>` | Download attachment | Low |
| `fm completion` | Generate shell completions | High (built-in with Cobra) |

## Authentication

Token is retrieved from 1Password:
```
op read "op://Services/Fastmail/credential"
```

Or via `FASTMAIL_TOKEN` environment variable.

Implement token caching (1 hour TTL) to avoid repeated Touch ID prompts.

## Development

```bash
# Build
go build -o fm ./cmd/fm

# Run
./fm inbox

# Install shell completions
./fm completion zsh > ~/.zsh/completions/_fm
```

## Implementation Plan

### Phase 1: Core Infrastructure
1. Set up Cobra root command with help and version
2. Implement JMAP client wrapper using go-jmap
3. Add authentication (1Password + env var + caching)
4. Implement `inbox` command as proof of concept

### Phase 2: Read Operations
5. `read` - Display full email
6. `thread` - Display conversation
7. `search` - Search with filters
8. `folders` - List mailboxes

### Phase 3: Write Operations
9. `archive` / `bulk-archive`
10. `move`
11. `draft` / `reply` / `forward`

### Phase 4: Polish
12. Shell completion with dynamic suggestions
13. Homebrew formula
14. Error handling and edge cases

## Releasing

When making meaningful changes, suggest creating a new release using semantic versioning:

- **PATCH** (v1.0.x): Bug fixes, documentation updates, minor tweaks
- **MINOR** (v1.x.0): New features, new commands, new flags (backwards compatible)
- **MAJOR** (vX.0.0): Breaking changes (renamed commands, changed flag behavior, removed features) - **always ask user for confirmation**

```bash
git tag v1.x.x
git push origin v1.x.x
```

This triggers GitHub Actions to build binaries and update the Homebrew tap.

## Reference

The original TypeScript implementation is in `node/src/`:
- `cli.ts` - Command parsing and output formatting
- `jmap-client.ts` - JMAP API calls
- `auth.ts` - Authentication

Use these as the specification for behavior and output format.
