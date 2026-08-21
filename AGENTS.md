# AGENTS.md

> **CRITICAL RULE**: ALL AI AGENTS MUST READ AND UNDERSTAND THIS FILE BEFORE EXECUTING OR MODIFYING ANY CODE IN THIS PROJECT. DO NOT PROCEED WITHOUT ADHERING TO THE RULES DEFINED BELOW.

## 1. Mandatory Activation Rules
- You MUST analyze the current state of this document before making any changes.
- Follow all coding styles, structures, and naming conventions strictly.
- Before committing any changes, you MUST update the "Current Project Status & Change History" section of this file to reflect the task progress and change documentation.

## 2. Context & Tech Stack Overview
This project is a Discord Bot written in Go, specifically designed to control, monitor, and manage a Minecraft server running on Linux via `systemctl`, `journalctl`, and `mcrcon`.

- **Language:** Go (Golang) version 1.21.2
- **Main Libraries:**
  - `github.com/bwmarrin/discordgo` (Discord Bot API)
  - `github.com/joho/godotenv` (Environment Variables)
- **External Dependencies & Tools:**
  - `systemctl` (to start/stop the Minecraft server service)
  - `journalctl` (to listen to server logs via stdout/stderr)
  - `mcrcon` (to send RCON commands like whitelist `/wl`)

## 3. Complete Directory Structure Map
```text
.
├── .env.example              # Template for environment variables
├── .gitignore                # Git ignored files configuration
├── Makefile                  # Build script and shorthand commands
├── Readme.md                 # Project documentation
├── go.mod                    # Go module dependencies
├── go.sum                    # Go module checksums
├── main.go                   # Application entry point, Discord bot initialization and routing
├── server/                   # Server configurations and deployment scripts
│   └── minecraft-systemd.service # Systemd service unit template for the Minecraft server
└── utils/                    # Helper packages and business logic modules
    ├── discord.go            # Discord command handlers (start, stop, wl)
    └── shell.go              # OS-level integrations (systemctl, journalctl tailing, regex)
```
**Rule for New Files:** Do not clutter the root directory. Place new features inside `utils/` or create new modular directories if needed.

## 4. Coding Rules & Constraints
- **Formatting:** Code MUST be formatted strictly using standard `gofmt`.
- **Modularity:** Main Discord API init and command registration belong in `main.go`. Business logic, event processing, and shell interactions belong in `utils/`.
- **Naming Conventions:** Use standard Go conventions (CamelCase for exported, camelCase for unexported).
- **Error Handling & Logging:** Use the standard `log` package (e.g., `log.Printf`, `log.Println`). Always handle errors properly.
- **Environment Variables:** All secrets and environment-specific settings (tokens, passwords, guild IDs) must be handled via `.env` and accessed through `os.Getenv()`.

## 5. Commands & Workflows
- **Build & Run Project:**
  - Run `make go` (Executes `go build .` and then `./minedc`).
- **Dependencies Management:**
  - `go mod tidy` (To clean and update dependencies).

## 6. Automatic Task Tracking & Change Documentation Rules
**MANDATORY FOR AI:** Every time you make changes, add code, or perform refactoring, you MUST:
1. Update the task status in the **Tasks Status** list below (change `[ ]` to `[/]` for in-progress, or `[x]` for completed). Add new tasks as necessary.
2. Add a new entry to the **Recent Changes / Changelog** section documenting what file was changed and the reasoning behind it.

## 7. Current Project Status & Change History

### Tasks Status
- [x] Added /backup, /tps, /logs, /ip, /op, /deop, and Cross-Chat
- [x] Created DISCORD_GUIDE.md
- [x] Added /info command for CPU and RAM monitoring
- [x] Added /status, /restart, /say, /kick, /ban commands
- [x] Initial Project Analysis & AGENTS.md Creation

- [x] Added /leaderboard command to display top player stats

- [x] Added /sync_titles command and automated LuckPerms title assignment for top players.

### Recent Changes / Changelog
- **2026-08-20**: `AGENTS.md` - Created comprehensive agent rule file based on project analysis to standardize AI interactions.
- **2026-08-20**: main.go, utils/discord.go - Added /status, /restart, /say, /kick, /ban commands via mcrcon and systemctl.
- **2026-08-20**: main.go, utils/discord.go - Added /info command for hardware monitoring.
- **2026-08-20**: Added advanced commands (/backup, /tps, /logs, /ip, /op, /deop) and Cross-Chat listener. Created DISCORD_GUIDE.md.- **2026-08-20**: main.go, utils/leaderboard.go - Added /leaderboard command to parse world/stats/ and display top player stats like deaths, kills, playtime, and diamonds mined.
- **2026-08-20**: main.go, utils/titles.go - Added auto-title assignment logic via RCON, connecting stats to LuckPerms.

## 8. Development & Debugging Access (VPS)
If direct server debugging is required, you can SSH into the production server using the following credentials:
- **IP Address:** 147.93.159.183
- **User:** minecraft
- **Password:** minecraft12345
- **System Architecture:** systemd (`mcbot` service for golang bot, `minecraft-server` service for the Minecraft server)
