package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	filePath := "/home/samsul/script/go/mcbot/main.go"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	strContent := string(content)

	cmdString := `{
			Name:        "leaderboard",
			Description: "View player statistics leaderboard",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "Select leaderboard category",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Paling Sering Mati (Deaths)", Value: "deaths"},
						{Name: "Raja PVP (Kills)", Value: "player_kills"},
						{Name: "Paling Lama Bermain", Value: "play_time"},
						{Name: "Penambang Diamond", Value: "diamonds"},
					},
				},
			},
		},`

	handlerString := `"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			category := ""
			for _, opt := range i.ApplicationCommandData().Options {
				if opt.Name == "category" {
					category = opt.StringValue()
				}
			}
			msg, err := utils.GenerateLeaderboard(category)
			if err != nil {
				msg = "Error: " + err.Error()
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msg,
				},
			})
		},`

	// Find commands = []*discordgo.ApplicationCommand{
	commandsIdx := strings.Index(strContent, `commands = []*discordgo.ApplicationCommand{`)
	if commandsIdx != -1 {
		insertIdx := commandsIdx + len(`commands = []*discordgo.ApplicationCommand{`)
		strContent = strContent[:insertIdx] + "\n\t\t" + cmdString + strContent[insertIdx:]
	}

	// Find commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	handlersIdx := strings.Index(strContent, `commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){`)
	if handlersIdx != -1 {
		insertIdx := handlersIdx + len(`commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){`)
		strContent = strContent[:insertIdx] + "\n\t\t" + handlerString + strContent[insertIdx:]
	}

	err = ioutil.WriteFile(filePath, []byte(strContent), 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	log.Println("Patch applied successfully")
}
