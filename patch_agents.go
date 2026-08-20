package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	filePath := "/home/samsul/script/go/mcbot/AGENTS.md"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	strContent := string(content)

	taskString := "- [x] Added /leaderboard command to display top player stats\n"
	changelogString := "- **2026-08-20**: main.go, utils/leaderboard.go - Added /leaderboard command to parse world/stats/ and display top player stats like deaths, kills, playtime, and diamonds mined.\n"

	// Find the end of Tasks Status
	tasksIdx := strings.Index(strContent, `### Recent Changes / Changelog`)
	if tasksIdx != -1 {
		strContent = strContent[:tasksIdx] + taskString + "\n" + strContent[tasksIdx:]
	}

	// Append to changelog
	strContent = strContent + changelogString

	err = ioutil.WriteFile(filePath, []byte(strContent), 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	log.Println("AGENTS.md patched successfully")
}
