package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type UserCache struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type PlayerStat struct {
	Name  string
	Score int
}

func getUUIDToNameMap(mcPath string) (map[string]string, error) {
	cachePath := filepath.Join(mcPath, "usercache.json")
	data, err := ioutil.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var users []UserCache
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	uuidMap := make(map[string]string)
	for _, u := range users {
		uuidMap[u.UUID] = u.Name
	}
	return uuidMap, nil
}

func getStatValue(statData map[string]interface{}, category string) int {
	stats, ok := statData["stats"].(map[string]interface{})
	if !ok {
		return 0
	}

	if category == "diamonds" {
		total := 0
		if group, exists := stats["minecraft:mined"].(map[string]interface{}); exists {
			if val, exists := group["minecraft:diamond_ore"].(float64); exists {
				total += int(val)
			}
			if val, exists := group["minecraft:deepslate_diamond_ore"].(float64); exists {
				total += int(val)
			}
		}
		return total
	}

	var statGroup, statName string
	switch category {
	case "deaths":
		statGroup = "minecraft:custom"
		statName = "minecraft:deaths"
	case "player_kills":
		statGroup = "minecraft:custom"
		statName = "minecraft:player_kills"
	case "play_time":
		statGroup = "minecraft:custom"
		statName = "minecraft:play_time"
	case "totems":
		statGroup = "minecraft:used"
		statName = "minecraft:totem_of_undying"
	}

	if group, exists := stats[statGroup].(map[string]interface{}); exists {
		if val, exists := group[statName].(float64); exists {
			return int(val)
		}
	}
	return 0
}

func GenerateLeaderboard(category string) (string, error) {
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

	var leaderboard []PlayerStat

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		uuid := strings.TrimSuffix(f.Name(), ".json")
		name, exists := uuidMap[uuid]
		if !exists {
			name = "Unknown (" + uuid[:8] + ")"
		}

		filePath := filepath.Join(statsDir, f.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read stat file %s: %v", f.Name(), err)
			continue
		}

		var statData map[string]interface{}
		if err := json.Unmarshal(data, &statData); err != nil {
			log.Printf("Failed to parse stat file %s: %v", f.Name(), err)
			continue
		}

		score := getStatValue(statData, category)
		if score > 0 {
			leaderboard = append(leaderboard, PlayerStat{Name: name, Score: score})
		}
	}

	if len(leaderboard) == 0 {
		return "No data found for this category.", nil
	}

	// Sort descending
	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].Score > leaderboard[j].Score
	})

	title := ""
	switch category {
	case "deaths":
		title = "💀 Paling Sering Mati"
	case "player_kills":
		title = "⚔️ Raja PVP (Player Kills)"
	case "play_time":
		title = "⏱️ Paling Lama Bermain (Ticks)"
	case "diamonds":
		title = "💎 Penambang Diamond Terbanyak"
	case "totems":
		title = "👼 Sembilan Nyawa (Pemakai Totem)"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**\n```text\n", title))
	sb.WriteString(fmt.Sprintf("%-5s | %-16s | %s\n", "Rank", "Player", "Score"))
	sb.WriteString(strings.Repeat("-", 35) + "\n")

	limit := 10
	if len(leaderboard) < limit {
		limit = len(leaderboard)
	}

	for i := 0; i < limit; i++ {
		scoreStr := fmt.Sprintf("%d", leaderboard[i].Score)
		if category == "play_time" {
			hours := (leaderboard[i].Score / 20) / 3600
			scoreStr = fmt.Sprintf("%dh", hours)
		}
		sb.WriteString(fmt.Sprintf("%-5d | %-16s | %s\n", i+1, leaderboard[i].Name, scoreStr))
	}
	sb.WriteString("```")

	return sb.String(), nil
}
