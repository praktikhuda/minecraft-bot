package utils

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func Systemctl(arg string, s *discordgo.Session, chName string) {
	servName := os.Getenv("SERVICE_NAME")
	cmd := exec.Command("sudo", "systemctl", arg, servName)
	cmd.Run()
}
func IsServiceRunning(serviceName string) bool {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()

	if err != nil {
		return false
	}

	status := strings.TrimSpace(string(output))
	return status == "active"
}

func LogListen(ctx context.Context, s *discordgo.Session, channelID string) {
	servName := os.Getenv("SERVICE_NAME")
	cmd := exec.CommandContext(ctx, "journalctl", "-u", servName, "-f")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error creating StdoutPipe:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting command:", err)
		return
	}
	log.Println("Listening Minecraft Server Log..")

	// Regex for server info lines
	// [12:34:56] [Server thread/INFO]: <MrPheee> hello
	chatRegex := regexp.MustCompile(`INFO\]: (?:\[Not Secure\] )?<([^>]+)> (.*)`)
	joinRegex := regexp.MustCompile(`INFO\]: ([a-zA-Z0-9_]+) joined the game`)
	leaveRegex := regexp.MustCompile(`INFO\]: ([a-zA-Z0-9_]+) left the game`)
	advancementRegex := regexp.MustCompile(`INFO\]: ([a-zA-Z0-9_]+ has made the advancement .*)`)
	notWhitelistRegex := regexp.MustCompile(`INFO\]: Disconnecting ([a-zA-Z0-9_]+) \(.*You are not white-listed`)
	detailedDeathRegex := regexp.MustCompile(`x=([-\d.]+), y=([-\d.]+), z=([-\d.]+).*?\] died, message: '(.*?)'`)

	// Death messages usually don't have < > and aren't standard join/leave/advancements.
	// But they are triggered when a player dies. We can match any line starting with a known player name
	// if it's not a chat message. To do this perfectly, we can match any line that contains words like "was slain", "fell", "blew up", "drowned", etc.
	// Or we can just catch lines matching `\[Server thread/INFO\]: ([a-zA-Z0-9_]+ .*)` and exclude server messages like "UUID", "logged in", "lost connection".
	
	deathKeywords := []string{
		"was shot by", "was pummeled by", "was pricked to death", "walked into a cactus", "drowned",
		"experienced kinetic energy", "blew up", "was blown up by", "was killed by", "hit the ground too hard",
		"fell from a high place", "fell off a ladder", "fell off some vines", "fell out of the water",
		"fell into a patch of fire", "fell into a patch of cacti", "went up in flames", "burned to death",
		"was burnt to a crisp", "went off with a bang", "tried to swim in lava", "was slain by",
		"was fireballed by", "was stung to death", "starved to death", "suffocated in a wall",
		"was squished too much", "was squashed by", "withered away", "died",
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping log listening...")
			return
		default:
			line := scanner.Text()

			var res string

			if match := chatRegex.FindStringSubmatch(line); match != nil {
				res = fmt.Sprintf("💬 **%s**: %s", match[1], match[2])
			} else if match := detailedDeathRegex.FindStringSubmatch(line); match != nil {
				res = parseDeathMessage(match[4], match[1], match[2], match[3])
			} else if match := joinRegex.FindStringSubmatch(line); match != nil {
				res = fmt.Sprintf("✅ **%s** joined the game", match[1])
			} else if match := leaveRegex.FindStringSubmatch(line); match != nil {
				res = fmt.Sprintf("❌ **%s** left the game", match[1])
			} else if match := advancementRegex.FindStringSubmatch(line); match != nil {
				res = fmt.Sprintf("🏆 **%s**", match[1])
			} else if match := notWhitelistRegex.FindStringSubmatch(line); match != nil {
				res = fmt.Sprintf("⚠️ **%s** tried to join but is not whitelisted.\nAdd player using:\n```/wl %s```", match[1], match[1])
			} else if strings.Contains(line, "For help, type") {
				res = "🟢 **Minecraft Server is Running!**"
				s.ChannelEdit(channelID, &discordgo.ChannelEdit{
					Name: "minecraft-on",
				})
			} else if strings.Contains(line, "INFO]:") {
				parts := strings.SplitN(line, "INFO]: ", 2)
				if len(parts) == 2 {
					msg := strings.TrimSpace(parts[1])
					if !strings.Contains(msg, "UUID of player") && !strings.Contains(msg, "logged in with entity id") && !strings.Contains(msg, "lost connection:") {
						for _, keyword := range deathKeywords {
							if strings.Contains(msg, keyword) {
								res = fmt.Sprintf("💀 **%s**", msg)
								break
							}
						}
					}
				}
			}

			if res != "" {
				fmt.Println("Sending to Discord:", res)
				s.ChannelMessageSend(channelID, res)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading log:", err)
	}

	cmd.Wait()
}


func parseDeathMessage(msg, x, y, z string) string {
	x = strings.Split(x, ".")[0]
	y = strings.Split(y, ".")[0]
	z = strings.Split(z, ".")[0]

	victim := ""
	killer := ""
	weapon := ""
	alasan := msg

	delimiters := []string{" was slain by ", " was shot by ", " was blown up by ", " was fireballed by ", " was pummeled by ", " was killed by "}
	
	for _, delim := range delimiters {
		if idx := strings.Index(msg, delim); idx != -1 {
			victim = msg[:idx]
			rest := msg[idx+len(delim):]
			if usingIdx := strings.Index(rest, " using "); usingIdx != -1 {
				killer = rest[:usingIdx]
				weapon = rest[usingIdx+7:]
				alasan = strings.TrimSpace(delim[4:])
			} else {
				killer = rest
				alasan = strings.TrimSpace(delim[4:])
			}
			break
		}
	}

	if victim == "" {
		parts := strings.SplitN(msg, " ", 2)
		if len(parts) == 2 {
			victim = parts[0]
			alasan = parts[1]
		} else {
			victim = "Unknown"
		}
	}

	var sb strings.Builder
	sb.WriteString("💀 **LAPORAN KEMATIAN** 💀\n")
	sb.WriteString(fmt.Sprintf("**Korban:** `%s`\n", victim))
	
	if killer != "" {
		sb.WriteString(fmt.Sprintf("**Pelaku:** `%s`\n", killer))
	}
	
	if weapon != "" {
		sb.WriteString(fmt.Sprintf("**Senjata:** `%s`\n", weapon))
	} else if killer == "" {
		sb.WriteString(fmt.Sprintf("**Penyebab:** `%s`\n", alasan))
	}

	sb.WriteString(fmt.Sprintf("**Lokasi (XYZ):** `X: %s, Y: %s, Z: %s`\n", x, y, z))
	
	return sb.String()
}
