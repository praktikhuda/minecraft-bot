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
		s.ChannelMessageSend(m.ChannelID, "⚙️ Memulai proses kompresi backup dunia (.tar.gz). Mohon tunggu, proses ini bisa memakan waktu tergantung ukuran dunia...")
		go func() {
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			backupFile := fmt.Sprintf("/tmp/backup_%s.tar.gz", timestamp)
			
			worldPath := os.Getenv("MINECRAFT_PATH")
			if worldPath == "" {
				worldPath = "/home/mcserv/server-minecraft"
			}
			
			cmd := exec.Command("sudo", "tar", "-czf", backupFile, "-C", worldPath, "world")
			err := cmd.Run()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ **Backup gagal (Kompresi error):** %v", err))
				return
			}
			
			// Ubah permission agar mcbot bisa membaca filenya untuk diupload
			exec.Command("sudo", "chmod", "644", backupFile).Run()
			
			s.ChannelMessageSend(m.ChannelID, "⏳ Kompresi selesai. Mulai mengunggah file backup ke Cloud (Uguu.se)...")
			
			uploadCmd := exec.Command("curl", "-s", "-F", "files[]=@"+backupFile, "https://uguu.se/upload.php")
			urlBytes, err := uploadCmd.Output()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ **Backup gagal (Upload error):** %v", err))
				return
			}
			
			// Parse JSON response dari Uguu.se
			type UguuResponse struct {
				Success bool `json:"success"`
				Files   []struct {
					URL string `json:"url"`
				} `json:"files"`
			}
			
			var result UguuResponse
			if err := json.Unmarshal(urlBytes, &result); err != nil || !result.Success || len(result.Files) == 0 {
				s.ChannelMessageSend(m.ChannelID, "❌ **Backup gagal (Server Cloud menolak file).**")
				return
			}
			
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ **Backup Berhasil!**\n\nKlik link di bawah ini untuk mengunduh dunia Anda:\n%s\n\n*(Link akan kedaluwarsa secara otomatis dalam 48 jam)*", result.Files[0].URL))
			
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
			cmds = append(cmds, fmt.Sprintf("op %s", username))
			cmds = append(cmds, fmt.Sprintf("gamemode spectator %s", username))
			actionName = "dijadikan SPECTATOR (Mata-mata & OP)"
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

		msg := sb.String()
			s.InteractionResponseEdit(m.Interaction, &discordgo.WebhookEdit{Content: &msg})

		case "leaderboard":
		if p == "" {
			s.ChannelMessageSend(m.ChannelID, "❌ **Error:** Kategori tidak ditemukan.")
			return
		}
		
		allCats := GetAllCategories()
		var tpetCat LBCategory
		found := false
		for _, c := range allCats {
			if c.StatID == p {
				tpetCat = c
				found = true
				break
			}
		}
		
		if !found {
			s.ChannelMessageSend(m.ChannelID, "❌ **Kategori tidak valid.**")
			return
		}
		
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("📊 **Sedang memuat statistik: %s...**", tpetCat.Name))
		
		go func() {
			results := GetLeaderboard(tpetCat.StatID)
			if len(results) == 0 {
				msg := "🤷‍♂️ **Belum ada data untuk kategori ini.**"
				s.InteractionResponseEdit(m.Interaction, &discordgo.WebhookEdit{Content: &msg})
				return
			}
			
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("🏆 **Leaderboard: %s**\n*%s*\n\n", tpetCat.Name, tpetCat.Desc))
			
			for i, r := range results {
				if i >= 10 { // Top 10
					break
				}
				medal := "🏅"
				if i == 0 { medal = "🥇" } else if i == 1 { medal = "🥈" } else if i == 2 { medal = "🥉" }
				sb.WriteString(fmt.Sprintf("%s **#%d %s** - %d\n", medal, i+1, r.Name, r.Value))
			}
			
			msg := sb.String()
			s.InteractionResponseEdit(m.Interaction, &discordgo.WebhookEdit{Content: &msg})
		}()

			case "gelar":
		if p == "" {
			s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ **Format Salah!** Gunakan: `/gelar [username]`",
				},
			})
			return
		}
		
		s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		
		go func() {
			titles := GetPlayerTitles(p)
			if len(titles) == 0 {
				msgText := fmt.Sprintf("😭 **%s tidak memiliki gelar apapun.**\n(Jadilah Rank 1 di kategori Leaderboard manapun untuk mendapatkan gelar!)", p)
				s.InteractionResponseEdit(m.Interaction, &discordgo.WebhookEdit{Content: &msgText})
				return
			}
			
			var options []discordgo.SelectMenuOption
			for _, t := range titles {
				starStr := strings.Repeat("⭐", t.Stars)
				safeName := strings.ReplaceAll(strings.ToLower(t.Name), " ", "_")
				options = append(options, discordgo.SelectMenuOption{
					Label:       fmt.Sprintf("%s %s", t.Title, starStr),
					Description: t.Name,
					Value:       fmt.Sprintf("%s:%s", p, safeName),
				})
			}
			
			// Add newbie title
			options = append(options, discordgo.SelectMenuOption{
				Label:       "[Warga Biasa]",
				Description: "Gelar Default",
				Value:       fmt.Sprintf("%s:newbie", p),
			})

			msgText := fmt.Sprintf("🎖️ **Gelar Eksklusif Milik %s** 🎖️\nSilakan pilih gelar yang ingin Anda pakai di dalam game:", p)
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "select_title",
							Placeholder: "Pilih gelar Anda...",
							Options:     options,
						},
					},
				},
			}
			_, err := s.InteractionResponseEdit(m.Interaction, &discordgo.WebhookEdit{
				Content: &msgText,
				Components: &components,
			})
			if err != nil {
				log.Println("Error sending components:", err)
			}
		}()

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
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	data := i.MessageComponentData()
	if len(data.Values) == 0 {
		return
	}

	val := data.Values[0]
	parts := strings.SplitN(val, ":", 2)
	if len(parts) != 2 {
		return
	}

	username := parts[0]
	category := parts[1]

	go func() {
		if category == "newbie" {
			RunRcon(fmt.Sprintf("team leave %s", username))
			msgText := fmt.Sprintf("Gelar untuk **%s** berhasil dilepas (kembali menjadi Warga Biasa).", username)
			emptyComponents := []discordgo.MessageComponent{}
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content:    &msgText,
				Components: &emptyComponents,
			})
			return
		}

		// Find the original title info
		var foundCat LBCategory
		found := false
		
		allCats := [][]LBCategory{LBMining, LBCombat, LBHunt, LBExplore, LBFood, LBFails, LBMisc, LBExtra}
		for _, catList := range allCats {
			for _, cat := range catList {
				safeName := strings.ReplaceAll(strings.ToLower(cat.Name), " ", "_")
				if safeName == category {
					foundCat = cat
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		teamName := "t_" + category
		if len(teamName) > 16 {
			// Older minecraft servers only allow 16 chars for team names
			teamName = teamName[:16]
		}

		RunRcon(fmt.Sprintf("team add %s", teamName))
		if found {
			RunRcon(fmt.Sprintf("team modify %s prefix \"[%s] \"", teamName, foundCat.Title))
		}
		out, err := sendRconCommand(fmt.Sprintf("team join %s %s", teamName, username))

		msgText := "Gelar berhasil dipasang!"
		if err != nil {
			msgText = fmt.Sprintf("Gagal memasang gelar: %v", err)
		} else {
			msgText = fmt.Sprintf("Gelar untuk **%s** berhasil diubah menjadi **%s**!\n*(%s)*", username, foundCat.Title, strings.TrimSpace(string(out)))
		}

		emptyComponents := []discordgo.MessageComponent{}
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content:    &msgText,
			Components: &emptyComponents,
		})
	}()
}
