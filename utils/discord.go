package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
)

var cancelFunc context.CancelFunc

// Helper function to send RCON commands
// Regex to match ANSI escape codes
var ansiRegex = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

// Helper function to send RCON commands
func sendRconCommand(command string) (string, error) {
	host := "127.0.0.1"
	port := os.Getenv("RCON_PORT")
	if port == "" {
		port = "25575" // fallback
	}
	password := os.Getenv("RCON_PASSWORD")

	cmd := exec.Command("mcrcon",
		"-H", host,
		"-P", port,
		"-p", password,
		command,
	)

	output, err := cmd.CombinedOutput()
	cleanOutput := ansiRegex.ReplaceAllString(string(output), "")
	return strings.TrimSpace(cleanOutput), err
}

func AutoStartLog(s *discordgo.Session) {
	servName := os.Getenv("SERVICE_NAME")
	syncChannel := os.Getenv("SYNC_CHANNEL_ID")
	if syncChannel != "" && IsServiceRunning(servName) && cancelFunc == nil {
		ctx, cancel := context.WithCancel(context.Background())
		cancelFunc = cancel
		go LogListen(ctx, s, syncChannel)
		log.Println("Auto-started LogListen on boot.")
	}
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
	case "backup":
		s.ChannelMessageSend(m.ChannelID, "⚙️ Memulai proses kompresi backup dunia (.tar.gz). Mohon tunggu, proses ini bisa memakan waktu...")
		go func() {
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			backupFile := fmt.Sprintf("/tmp/backup_%s.tar.gz", timestamp)
			
			worldPath := os.Getenv("MINECRAFT_PATH")
			if worldPath == "" {
				worldPath = "/home/minecraft/server-minecraft"
			}
			
			cmd := exec.Command("sudo", "tar", "-czf", backupFile, "-C", worldPath, "world")
			err := cmd.Run()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ **Backup gagal (Kompresi error):** %v", err))
				return
			}
			
			// Change permission so bot can read it
			exec.Command("sudo", "chmod", "644", backupFile).Run()
			
			file, err := os.Open(backupFile)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "❌ **Gagal membaca file backup dari penyimpanan.**")
				return
			}
			defer file.Close()
			
			stat, err := file.Stat()
			if err == nil && stat.Size() > 24*1024*1024 {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ **Ukuran backup (%.2f MB) melebihi batas maksimal Discord (24 MB)!**\nFile tersimpan di VPS: `%s`", float64(stat.Size())/1024/1024, backupFile))
				return
			}
			
			s.ChannelMessageSend(m.ChannelID, "⏳ Kompresi selesai. Mengunggah file langsung ke Discord...")
			
			_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Content: "✅ **Backup Berhasil!**\nIni adalah file ZIP dunia Minecraft Anda:",
				Files: []*discordgo.File{
					{
						Name:        fmt.Sprintf("mc_world_backup_%s.tar.gz", timestamp),
						ContentType: "application/gzip",
						Reader:      file,
					},
				},
			})
			
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ **Gagal mengirim file ke Discord:** %v", err))
			}
			
			exec.Command("sudo", "rm", "-f", backupFile).Run()
		}()
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
	case "unban":
		if !IsServiceRunning(servName) {
			s.ChannelMessageSend(m.ChannelID, "Server is offline.")
			return
		}
		output, err := sendRconCommand(fmt.Sprintf("pardon %s", p))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to unban player.")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Result: %s", output))
		}
	case "op":
		ownerID := os.Getenv("OWNER_ID")
		userID := ""
		if m.Member != nil {
			userID = m.Member.User.ID
		} else if m.User != nil {
			userID = m.User.ID
		}
		if userID != ownerID {
			s.ChannelMessageSend(m.ChannelID, "❌ Akses Ditolak: Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!")
			return
		}
		if !IsServiceRunning(servName) {
			s.ChannelMessageSend(m.ChannelID, "Server is offline.")
			return
		}

		parts := strings.Split(p, ":")
		if len(parts) != 2 {
			// Jika format payload lama atau salah
			output, _ := sendRconCommand(fmt.Sprintf("op %s", p))
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Result: %s", output))
			return
		}
		username := parts[0]
		role := parts[1]

		var cmds []string
		var actionName string
		if role == "admin" {
			cmds = append(cmds, fmt.Sprintf("op %s", username))
			cmds = append(cmds, fmt.Sprintf("gamemode creative %s", username))
			actionName = "dijadikan ADMIN (OP & Creative)"
		} else if role == "spectator" {
			cmds = append(cmds, fmt.Sprintf("deop %s", username))
			cmds = append(cmds, fmt.Sprintf("gamemode spectator %s", username))
			actionName = "dijadikan SPECTATOR (Mata-mata)"
		} else if role == "player" {
			cmds = append(cmds, fmt.Sprintf("deop %s", username))
			cmds = append(cmds, fmt.Sprintf("gamemode survival %s", username))
			actionName = "dikembalikan menjadi PLAYER (Survival & Deop)"
		}

		var finalMsg string
		for _, cmd := range cmds {
			_, err := sendRconCommand(cmd)
			if err != nil {
				finalMsg = "❌ Error mengeksekusi RCON: " + err.Error()
				break
			}
		}

		if finalMsg == "" {
			finalMsg = fmt.Sprintf("✅ **Berhasil:** `%s` telah %s!", username, actionName)
		}

		s.ChannelMessageSend(m.ChannelID, finalMsg)
	case "deop":
		ownerID := os.Getenv("OWNER_ID")
		userID := ""
		if m.Member != nil {
			userID = m.Member.User.ID
		} else if m.User != nil {
			userID = m.User.ID
		}
		if userID != ownerID {
			s.ChannelMessageSend(m.ChannelID, "❌ Akses Ditolak: Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!")
			return
		}
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
	case "whereis":
		ownerID := os.Getenv("OWNER_ID")
		userID := ""
		if m.Member != nil {
			userID = m.Member.User.ID
		} else if m.User != nil {
			userID = m.User.ID
		}
		if userID != ownerID {
			s.ChannelMessageSend(m.ChannelID, "❌ Akses Ditolak: Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!")
			return
		}
		if !IsServiceRunning(servName) {
			s.ChannelMessageSend(m.ChannelID, "Server is offline.")
			return
		}

		outList, err := sendRconCommand("list")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to get player list.")
			return
		}

		// Parse list output: "There are X of a max of Y players online: name1, name2"
		listStr := string(outList)
		idx := strings.Index(listStr, "online: ")
		if idx == -1 {
			s.ChannelMessageSend(m.ChannelID, "Tidak ada pemain yang online.")
			return
		}

		playersPart := listStr[idx+8:]
		if len(strings.TrimSpace(playersPart)) == 0 {
			s.ChannelMessageSend(m.ChannelID, "Tidak ada pemain yang online.")
			return
		}


		players := strings.Split(playersPart, ", ")

		if p != "" {
			found := false
			for _, pl := range players {
				if strings.EqualFold(strings.TrimSpace(pl), strings.TrimSpace(p)) {
					players = []string{pl}
					found = true
					break
				}
			}
			if !found {
				s.ChannelMessageSend(m.ChannelID, "Pemain **" + p + "** sedang tidak online atau tidak ditemukan.")
				return
			}
		}

		var sb strings.Builder

		sb.WriteString(fmt.Sprintf("**📍 Live Player Locations (%d Online):**\n\n", len(players)))

		for _, p := range players {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}

			// Pos
			posOut, _ := sendRconCommand(fmt.Sprintf("data get entity %s Pos", p))
			posStr := string(posOut)
			// expected: name has the following entity data: [X.Xd, Y.Yd, Z.Zd]
			startB := strings.Index(posStr, "[")
			endB := strings.Index(posStr, "]")

			coords := "Unknown"
			if startB != -1 && endB != -1 {
				rawCoords := posStr[startB+1 : endB]
				parts := strings.Split(rawCoords, ", ")
				if len(parts) == 3 {
					x := strings.TrimSuffix(parts[0], "d")
					y := strings.TrimSuffix(parts[1], "d")
					z := strings.TrimSuffix(parts[2], "d")
					// Convert floats to int display if possible, or just slice
					x = strings.Split(x, ".")[0]
					y = strings.Split(y, ".")[0]
					z = strings.Split(z, ".")[0]
					coords = fmt.Sprintf("X: %s, Y: %s, Z: %s", x, y, z)
				}
			}

			// Dimension
			dimOut, _ := sendRconCommand(fmt.Sprintf("data get entity %s Dimension", p))
			dimStr := string(dimOut)

			dim := "Unknown"
			if strings.Contains(dimStr, "\"minecraft:overworld\"") {
				dim = "🌳 Overworld"
			} else if strings.Contains(dimStr, "\"minecraft:the_nether\"") {
				dim = "🔥 Nether"
			} else if strings.Contains(dimStr, "\"minecraft:the_end\"") {
				dim = "🌌 The End"
			}

			sb.WriteString(fmt.Sprintf("- **%s**: %s (%s)\n", p, dim, coords))
		}

		s.ChannelMessageSend(m.ChannelID, sb.String())

	case "listleaderboard":
		var sb strings.Builder
		sb.WriteString("🏆 **Daftar Gelar (Leaderboard) Tersedia** 🏆\n\n")

		grouped := make(map[string][]string)
		for catKey, info := range Titles {
			if catKey == "newbie" {
				continue
			}
			if p != "" && !strings.EqualFold(info.Category, p) {
				continue
			}
			grouped[info.Category] = append(grouped[info.Category], fmt.Sprintf("**%s**\n└ Cara dapat: *%s*", info.PlainName, info.Desc))
		}

		if len(grouped) == 0 {
			s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Kategori tidak ditemukan atau belum ada gelar di kategori ini."},
			})
			return
		}

		for catName, items := range grouped {
			sb.WriteString(fmt.Sprintf("📂 **Kategori: %s**\n", strings.ToUpper(catName)))
			for _, item := range items {
				sb.WriteString(item + "\n")
			}
			sb.WriteString("\n")
		}

		s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: sb.String()},
		})

	case "gelar":
		if p == "" {
			s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Username tidak valid."},
			})
			return
		}

		ownedTitles, err := GetTitlesForPlayer(p)
		if err != nil {
			s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Gagal mengambil data gelar."},
			})
			return
		}

		if len(ownedTitles) == 0 {
			runRcon(fmt.Sprintf("team add title_newbie"))
			runRcon(fmt.Sprintf("team modify title_newbie suffix {\"text\":\" [Warga Biasa]\",\"color\":\"gray\"}"))
			runRcon(fmt.Sprintf("team join title_newbie %s", p))
			s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Maaf, pemain **%s** belum memenangkan gelar bergengsi apapun.\nGelar Anda saat ini diatur menjadi: **[Warga Biasa]**", p),
				},
			})
			return
		}

		var options []discordgo.SelectMenuOption
		for cat, tInfo := range ownedTitles {
			val := fmt.Sprintf("%s:%s", p, cat)
			options = append(options, discordgo.SelectMenuOption{
				Label:       tInfo.PlainName,
				Description: fmt.Sprintf("Gunakan gelar %s", tInfo.PlainName),
				Value:       val,
			})
		}

		s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("🎓 **Gelar milik %s**\nSilakan pilih gelar mana yang ingin Anda pakai di dalam game:", p),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    "select_gelar",
								Placeholder: "Pilih gelar Anda...",
								Options:     options,
							},
						},
					},
				},
			},
		})

	case "wllist":
		mcPath := os.Getenv("MINECRAFT_PATH")
		wlPath := filepath.Join(mcPath, "whitelist.json")
		data, err := ioutil.ReadFile(wlPath)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Gagal membaca whitelist.json")
			return
		}
		var wl []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &wl); err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Gagal mem-parsing whitelist.json")
			return
		}
		var names []string
		for _, v := range wl {
			names = append(names, "- "+v.Name)
		}
		msg := fmt.Sprintf("**📋 Daftar Pemain Whitelist (%%d):**\n```text\n%%s\n```", len(names), strings.Join(names, "\n"))
		s.ChannelMessageSend(m.ChannelID, msg)
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


func HandleGelarSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	if len(data.Values) == 0 {
		return
	}

	val := data.Values[0]
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return
	}

	username := parts[0]
	category := parts[1]

	teamName := "title_" + category

	// Create the team if it doesn't exist just in case
	runRcon(fmt.Sprintf("team add %s", teamName))
	if tInfo, ok := Titles[category]; ok {
		runRcon(fmt.Sprintf("team modify %s suffix %s", teamName, tInfo.JSONSuffix))
	}

	out, err := sendRconCommand(fmt.Sprintf("team join %s %s", teamName, username))

	msg := "Gelar berhasil dipasang!"
	if err != nil {
		msg = fmt.Sprintf("Gagal memasang gelar: %v", err)
	} else {
		plainName := ""
		if tInfo, ok := Titles[category]; ok {
			plainName = tInfo.PlainName
		}
		msg = fmt.Sprintf("Gelar untuk **%s** berhasil diubah menjadi **%s**!\n*(%s)*", username, plainName, strings.TrimSpace(string(out)))
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    msg,
			Components: []discordgo.MessageComponent{},
		},
	})
}
