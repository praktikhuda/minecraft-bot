package utils

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"os"

	"github.com/bwmarrin/discordgo"
)

var cancelFunc context.CancelFunc

// Helper function to send RCON commands
func sendRconCommand(command string) (string, error) {
	host := "127.0.0.1"
	port := "25575"
	password := os.Getenv("RCON_PASSWORD")

	cmd := exec.Command("mcrcon",
		"-H", host,
		"-P", port,
		"-p", password,
		command,
	)

	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func MessageHandler(command string, s *discordgo.Session, m *discordgo.InteractionCreate, p string) {

	servName := os.Getenv("SERVICE_NAME")

	switch command {
	case "start":
		if IsServiceRunning(servName) {
			msg := "Minecraft Server is already Running!"
			s.ChannelMessageSend(m.ChannelID, msg)
			log.Println(msg)
			if cancelFunc == nil {
				ctx, cancel := context.WithCancel(context.Background())
				cancelFunc = cancel
				go LogListen(ctx, s, m.ChannelID)
			}
		} else {
			startServer()
			if cancelFunc != nil {
				fmt.Println("Log already running")
				return
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancelFunc = cancel
			go LogListen(ctx, s, m.ChannelID)
		}
	case "stop":
		var msg string
		if IsServiceRunning(servName) {
			msg = "**Minecraft Server is Stopped!**"
			cmd := exec.Command("sudo", "systemctl", "stop", servName)
			cmd.Run()
			log.Println("Server Service is stopped.")
			if cancelFunc != nil {
				cancelFunc()
				cancelFunc = nil
				fmt.Println("Log stopped")
			} else {
				fmt.Println("Log is not running")
			}
		} else {
			msg = "Minecraft Server is not Runnning!"
		}
		s.ChannelMessageSend(m.ChannelID, msg)
		log.Println("Message: ", msg)
		s.ChannelEdit(m.ChannelID, &discordgo.ChannelEdit{
			Name: "minecraft-off",
		})
		log.Println("Channel name updated.")
	case "restart":
		var msg string
		if IsServiceRunning(servName) {
			msg = "**Minecraft Server is Restarting...**"
			s.ChannelMessageSend(m.ChannelID, msg)
			cmd := exec.Command("sudo", "systemctl", "restart", servName)
			cmd.Run()
			log.Println("Server Service is restarted.")
		} else {
			msg = "Minecraft Server is not Runnning! Use `/start` instead."
			s.ChannelMessageSend(m.ChannelID, msg)
		}
	case "status":
		if IsServiceRunning(servName) {
			output, err := sendRconCommand("list")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Minecraft Server is active, but unable to reach RCON. Server might be still starting.")
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Minecraft Server Status:** Online 🟢\n**Players:** %s", output))
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "**Minecraft Server Status:** Offline 🔴")
		}
	case "info":
		// Get CPU Usage
		cmdCpu := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | awk '{print $2 + $4}'")
		cpuOut, _ := cmdCpu.Output()
		cpuUsage := strings.TrimSpace(string(cpuOut))
		if cpuUsage == "" {
			cpuUsage = "N/A"
		}

		// Get RAM Usage
		cmdRam := exec.Command("sh", "-c", "free -m | awk 'NR==2{printf \"%.2f%% (%sMB / %sMB)\", $3*100/$2, $3, $2 }'")
		ramOut, _ := cmdRam.Output()
		ramUsage := strings.TrimSpace(string(ramOut))
		if ramUsage == "" {
			ramUsage = "N/A"
		}

		// Get Uptime
		cmdUptime := exec.Command("uptime", "-p")
		uptimeOut, _ := cmdUptime.Output()
		uptime := strings.TrimSpace(string(uptimeOut))

		response := fmt.Sprintf("**Server Hardware Info:**\n💻 **CPU Usage:** `%s%%`\n🧠 **RAM Usage:** `%s`\n⏱ **Uptime:** `%s`", cpuUsage, ramUsage, uptime)

		s.ChannelMessageSend(m.ChannelID, response)
	case "tps":
		if !IsServiceRunning(servName) {
			s.ChannelMessageSend(m.ChannelID, "Server is offline.")
			return
		}
		// Try spark tps first
		output, err := sendRconCommand("spark tps")
		if err != nil || strings.Contains(output, "Unknown command") || strings.Contains(output, "Usage:") {
			output, _ = sendRconCommand("tps")
		}
		if output == "" {
			output = "Failed to retrieve TPS. Spark or TPS command might not be supported on this server."
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Server TPS:**\n```text\n%s\n```", output))
	case "ip":
		cmdIp := exec.Command("curl", "-s", "ifconfig.me")
		ipOut, _ := cmdIp.Output()
		publicIP := strings.TrimSpace(string(ipOut))
		if publicIP == "" {
			publicIP = "Unknown IP"
		}

		embed := &discordgo.MessageEmbed{
			Title:       "🌐 Server Connection Info",
			Description: fmt.Sprintf("You can join the server using the IP below:\n\n**IP Address:** `%s`\n**Version:** (Cek ke Admin)\n\n*Have fun playing!*", publicIP),
			Color:       0x00FF00,
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
	case "backup":
		s.ChannelMessageSend(m.ChannelID, "⚙️ Executing backup task in the background. I will notify you when it's done.")
		go func() {
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			backupFile := fmt.Sprintf("/root/backup_%s.tar.gz", timestamp)
			cmd := exec.Command("sudo", "tar", "-czf", backupFile, "/root/world")
			err := cmd.Run()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ **Backup failed:** %v", err))
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ **Backup successful!**\nFile saved as: `%s`", backupFile))
			}
		}()
	case "logs":
		cmdLogs := exec.Command("sudo", "journalctl", "-u", servName, "-n", "15", "--no-pager")
		logsOut, _ := cmdLogs.Output()
		logStr := strings.TrimSpace(string(logsOut))
		if logStr == "" {
			logStr = "No logs found."
		}
		// Ensure it doesn't exceed Discord limit
		if len(logStr) > 1900 {
			logStr = logStr[len(logStr)-1900:]
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Last 15 lines of server logs:**\n```text\n%s\n```", logStr))
	case "say":
		if !IsServiceRunning(servName) {
			s.ChannelMessageSend(m.ChannelID, "Server is offline.")
			return
		}
		output, err := sendRconCommand(fmt.Sprintf("say %s", p))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to send message.")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Broadcast sent: %s", output))
		}
	case "kick":
		if !IsServiceRunning(servName) {
			s.ChannelMessageSend(m.ChannelID, "Server is offline.")
			return
		}
		output, err := sendRconCommand(fmt.Sprintf("kick %s", p))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to kick player.")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Result: %s", output))
		}
	case "ban":
		if !IsServiceRunning(servName) {
			s.ChannelMessageSend(m.ChannelID, "Server is offline.")
			return
		}
		output, err := sendRconCommand(fmt.Sprintf("ban %s", p))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to ban player.")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Result: %s", output))
		}
	case "op":
		if !IsServiceRunning(servName) {
			s.ChannelMessageSend(m.ChannelID, "Server is offline.")
			return
		}
		output, err := sendRconCommand(fmt.Sprintf("op %s", p))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to OP player.")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Result: %s", output))
		}
	case "deop":
		if !IsServiceRunning(servName) {
			s.ChannelMessageSend(m.ChannelID, "Server is offline.")
			return
		}
		output, err := sendRconCommand(fmt.Sprintf("deop %s", p))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to De-OP player.")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Result: %s", output))
		}
	case "wl":
		output, err := sendRconCommand(fmt.Sprintf("whitelist add %s", p))
		if err != nil {
			log.Printf("Error executing command: %s\n", err)
			return
		}
		s.ChannelMessageSend(m.ChannelID, string(output))

	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown command")
	}
}

func startServer() {
	servName := os.Getenv("SERVICE_NAME")
	cmd := exec.Command("sudo", "systemctl", "start", servName)
	cmd.Run()
	log.Println("Starting Server Service.")
}

func CrossChatHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	syncChannel := os.Getenv("SYNC_CHANNEL_ID")

	// Ignore if no sync channel is set
	if syncChannel == "" {
		return
	}
	// Ignore if the message isn't in the sync channel
	if m.ChannelID != syncChannel {
		return
	}
	// Ignore messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	servName := os.Getenv("SERVICE_NAME")
	if !IsServiceRunning(servName) {
		return
	}

	// Clean up message
	content := strings.ReplaceAll(m.Content, "`", "'")
	content = strings.ReplaceAll(content, "\n", " ")

	// Send to Minecraft via RCON say
	sendRconCommand(fmt.Sprintf("say [DC - %s]: %s", m.Author.Username, content))
}
