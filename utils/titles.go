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

func runRcon(cmdStr string) error {
	rconPass := os.Getenv("RCON_PASSWORD")
	cmd := exec.Command("mcrcon", "-H", "127.0.0.1", "-p", rconPass, cmdStr)
	err := cmd.Run()
	if err != nil {
		log.Printf("RCON failed for %s: %v", cmdStr, err)
	}
	return err
}

func SyncTitles() (string, error) {
	mcPath := os.Getenv("MINECRAFT_PATH")
	if mcPath == "" {
		return "", fmt.Errorf("MINECRAFT_PATH is not set in .env")
	}

	uuidMap, err := getUUIDToNameMap(mcPath)
	if err != nil {
		return "", fmt.Errorf("Failed to read usercache.json: %v", err)
	}

	statsDir := filepath.Join(mcPath, "world", "players", "stats")
	files, err := ioutil.ReadDir(statsDir)
	if err != nil {
		return "", fmt.Errorf("Failed to read stats directory: %v", err)
	}

	// Maps to hold max score and the winner's name for each category
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

	// 1. Find the top players
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

	// 2. Clear old suffixes from EVERYONE (using weight 100)
	for _, name := range uuidMap {
		if name != "Unknown" {
			runRcon(fmt.Sprintf("lp user %s meta removesuffix 100", name))
		}
	}

	// 3. Assign new suffixes
	titles := map[string]string{
		"deaths":       " &c[Si Paling Tumbal]",
		"player_kills": " &4[Raja PVP]",
		"play_time":    " &e[Sepuh Server]",
		"diamonds":     " &b[Juragan Diamond]",
		"totems":       " &a[Sembilan Nyawa]",
	}

	var sb strings.Builder
	sb.WriteString("**✅ Sinkronisasi Gelar (Titles) Berhasil!**\n\n")

	for cat, win := range winners {
		if win.Name != "" && win.Score > 0 {
			suffix := titles[cat]
			cmdStr := fmt.Sprintf("lp user %s meta addsuffix 100 \"%s\"", win.Name, suffix)
			runRcon(cmdStr)

			sb.WriteString(fmt.Sprintf("**%s** -> %s (Score: %d)\n", win.Name, strings.ReplaceAll(suffix, "&", ""), win.Score))
		}
	}

	return sb.String(), nil
}
