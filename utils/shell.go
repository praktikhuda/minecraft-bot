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
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping log listening...")
			return
		default:
			line := scanner.Text()

			CurrentTourney.Mutex.Lock()
			isActive := CurrentTourney.IsActive
			CurrentTourney.Mutex.Unlock()
			if isActive {
				go HandleKillEvent(line)
			}

			if strings.Contains(line, "#msg") || strings.Contains(line, "joined the game") || strings.Contains(line, "left the game") || strings.Contains(line, "For help, type") || strings.Contains(line, "You are not white-listed on this server") {
				fmt.Println("server: ", line)
				msg := strings.Split(line, ":")
				tmp := msg[len(msg)-1]
				var res string
				log.Printf("message is: %v", line)
				if strings.Contains(line, "joined the game") {
					name := strings.TrimSpace(strings.TrimSuffix(tmp, " joined the game"))
					res = fmt.Sprintf("%s: **joined the game**", name)
				} else if strings.Contains(line, "For help, type") {
					res = "**Minecraft Server is Running!**"
					s.ChannelEdit(channelID, &discordgo.ChannelEdit{
						Name: "minecraft-on",
					})
				} else if strings.Contains(line, "left the game") {
					name := strings.TrimSpace(strings.TrimSuffix(tmp, " left the game"))
					res = fmt.Sprintf("%s: **left the game**", name)
				} else if strings.Contains(line, "#msg") {
					start := strings.Index(tmp, "<") + 1
					end := strings.Index(tmp, ">")
					username := tmp[start:end]
					message := strings.TrimPrefix(tmp[end+2:], "#msg ")
					res = fmt.Sprintf("%s: *%s*", username, message)
				} else if strings.Contains(line, "You are not white-listed") {
					re := regexp.MustCompile(`Disconnecting (.+?) \(`)
					matches := re.FindStringSubmatch(line)
					if len(matches) == 2 {
						name := matches[1]
						res = fmt.Sprintf(
							"**%v** tried to join but is not whitelisted.\nAdd player using:\n```/wl %v```",
							name,
							name,
						)
					}
				}
				s.ChannelMessageSend(channelID, res)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading log:", err)
	}

	cmd.Wait()
}
