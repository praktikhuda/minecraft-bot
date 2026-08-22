package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TeamInfo struct {
	JSONSuffix string
	PlainName  string
}

var Titles = map[string]TeamInfo{
	"deaths":       {"{\"text\":\" [Si Paling Tumbal]\",\"color\":\"red\"}", "[Si Paling Tumbal]"},
	"player_kills": {"{\"text\":\" [Raja PVP]\",\"color\":\"dark_red\"}", "[Raja PVP]"},
	"play_time":    {"{\"text\":\" [Sepuh Server]\",\"color\":\"yellow\"}", "[Sepuh Server]"},
	"diamonds":     {"{\"text\":\" [Juragan Diamond]\",\"color\":\"aqua\"}", "[Juragan Diamond]"},
	"totems":       {"{\"text\":\" [Sembilan Nyawa]\",\"color\":\"green\"}", "[Sembilan Nyawa]"},
	"newbie":       {"{\"text\":\" [Warga Biasa]\",\"color\":\"gray\"}", "[Warga Biasa]"},
}

func runRcon(cmdStr string) error {
	rconPass := os.Getenv("RCON_PASSWORD")
	cmd := exec.Command("mcrcon", "-H", "127.0.0.1", "-p", rconPass, cmdStr)
	err := cmd.Run()
	if err != nil {
		log.Printf("RCON failed for %s: %v", cmdStr, err)
	}
	return err
}

func CalculateWinners() (map[string]string, error) {
	mcPath := os.Getenv("MINECRAFT_PATH")
	if mcPath == "" {
		return nil, fmt.Errorf("MINECRAFT_PATH is not set in .env")
	}

	uuidMap, err := getUUIDToNameMap(mcPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read usercache.json: %v", err)
	}

	statsDir := filepath.Join(mcPath, "world", "players", "stats")
	files, err := ioutil.ReadDir(statsDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to read stats directory: %v", err)
	}

	type winner struct {
		Name  string
		Score int
	}
	winners := map[string]*winner{
		"deaths":       {Name: "", Score: 0},
		"player_kills": {Name: "", Score: 0},
		"play_time":    {Name: "", Score: 0},
		"diamonds":     {Name: "", Score: 0},
		"totems":       {Name: "", Score: 0},
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		uuid := strings.TrimSuffix(f.Name(), ".json")
		name, exists := uuidMap[uuid]
		if !exists {
			name = "Unknown"
			continue
		}

		filePath := filepath.Join(statsDir, f.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue
		}

		var statData map[string]interface{}
		if err := json.Unmarshal(data, &statData); err != nil {
			continue
		}

		for cat, win := range winners {
			score := getStatValue(statData, cat)
			if score > win.Score {
				win.Score = score
				win.Name = name
			}
		}
	}

	result := make(map[string]string)
	for cat, win := range winners {
		if win.Name != "" && win.Score > 0 {
			result[cat] = win.Name
		}
	}
	return result, nil
}

func GetTitlesForPlayer(username string) (map[string]TeamInfo, error) {
	winners, err := CalculateWinners()
	if err != nil {
		return nil, err
	}

	owned := make(map[string]TeamInfo)
	for cat, winnerName := range winners {
		if strings.EqualFold(winnerName, username) {
			owned[cat] = Titles[cat]
		}
	}
	return owned, nil
}

func SetupTeams() {
	for cat, tInfo := range Titles {
		teamName := "title_" + cat
		runRcon(fmt.Sprintf("team add %s", teamName))
		runRcon(fmt.Sprintf("team modify %s suffix %s", teamName, tInfo.JSONSuffix))
	}
}

func SyncTitles() (string, error) {
	winners, err := CalculateWinners()
	if err != nil {
		return "", err
	}

	SetupTeams()

	for cat := range Titles {
		if cat == "newbie" {
			continue // Don't empty newbie team on sync, let them keep it
		}
		teamName := "title_" + cat
		runRcon(fmt.Sprintf("team empty %s", teamName)) // Hapus semua pemain dari tim ini agar di-reset
	}

	var sb strings.Builder
	sb.WriteString("**✅ Sinkronisasi Gelar (Titles) Berhasil!**\n\n")

	for cat, name := range winners {
		teamName := "title_" + cat
		runRcon(fmt.Sprintf("team join %s %s", teamName, name))

		plainName := Titles[cat].PlainName
		sb.WriteString(fmt.Sprintf("**%s** -> %s\n", name, plainName))
	}

	return sb.String(), nil
}
