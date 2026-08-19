package utils

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

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