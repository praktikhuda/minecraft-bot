package main

import (
	"flag"
	"log"
	"fmt"
	"minedc/utils"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	GuildID        *string
	BotToken       *string
	RemoveCommands *bool
	s              *discordgo.Session
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}
}

func init() {
	GuildID = flag.String("guild", os.Getenv("GUILD_ID"), "Test guild ID. If not passed - bot registers commands globally")
	BotToken = flag.String("token", os.Getenv("BOT_TOKEN"), "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
	flag.Parse()

	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	integerOptionMinValue         = 1.0
	dmPermission                  = false
	defaultMemberPermisions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "start",
			Description: "Start Minecraft Server",
		},
		{
			Name:        "stop",
			Description: "Stop Minecraft Server",
		},
		{
			Name:        "restart",
			Description: "Restart Minecraft Server",
		},
		{
			Name:        "status",
			Description: "Check Minecraft Server Status and Players",
		},
		{
			Name:        "info",
			Description: "Check Server Hardware (CPU/RAM) Status",
		},
		{
			Name:        "wl",
			Description: "Whitelist a user on the Minecraft server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "The Minecraft username to whitelist",
					Required:    true,
				},
			},
		},
		{
			Name:        "say",
			Description: "Broadcast a message to the Minecraft server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "The message to broadcast",
					Required:    true,
				},
			},
		},
		{
			Name:        "kick",
			Description: "Kick a player from the Minecraft server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "The Minecraft username to kick",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "reason",
					Description: "The reason for kicking",
					Required:    false,
				},
			},
		},
		{
			Name:        "ban",
			Description: "Ban a player from the Minecraft server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "The Minecraft username to ban",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"start": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Starting Minecraft Server...**",
				},
			})
			utils.MessageHandler("start", s, i, "")
		},
		"stop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Stopping Minecraft server...**",
				},
			})
			utils.MessageHandler("stop", s, i, "")
		},
		"restart": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Restarting Minecraft server...**",
				},
			})
			utils.MessageHandler("restart", s, i, "")
		},
		"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Checking server status...**",
				},
			})
			utils.MessageHandler("status", s, i, "")
		},
		"info": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Gathering server info...**",
				},
			})
			utils.MessageHandler("info", s, i, "")
		},
		"say": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			message := ""
			for _, opt := range options {
				if opt.Name == "message" {
					message = opt.StringValue()
				}
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Broadcasting message...**",
				},
			})
			utils.MessageHandler("say", s, i, message)
		},
		"kick": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			username := ""
			reason := ""
			for _, opt := range options {
				if opt.Name == "username" {
					username = opt.StringValue()
				} else if opt.Name == "reason" {
					reason = opt.StringValue()
				}
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("**Kicking `%s`...**", username),
				},
			})
			arg := username
			if reason != "" {
				arg = fmt.Sprintf("%s %s", username, reason)
			}
			utils.MessageHandler("kick", s, i, arg)
		},
		"ban": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			username := ""
			for _, opt := range options {
				if opt.Name == "username" {
					username = opt.StringValue()
				}
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("**Banning `%s`...**", username),
				},
			})
			utils.MessageHandler("ban", s, i, username)
		},
		"wl": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			var username string
			if opt, ok := optionMap["username"]; ok {
				username = opt.StringValue()
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("**Adding `%s` to whitelist...**", username),
				},
			})
			utils.MessageHandler("wl", s, i, username)
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")

		commandsGlo, err := s.ApplicationCommands(s.State.User.ID, "")
		if err != nil {
			log.Fatalf("Cannot fetch guild commands: %v", err)
		}
		commandsGui, err := s.ApplicationCommands(s.State.User.ID, *GuildID)
		if err != nil {
			log.Fatalf("Cannot fetch guild commands: %v", err)
		}
		log.Printf("CMD: %v - %v", commandsGlo, commandsGui)
		for _, cmd := range commandsGlo {
			err := s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
			if err != nil {
				log.Printf("Cannot delete global command %s: %v", cmd.Name, err)
			} else {
				log.Printf("Deleted global command: %s", cmd.Name)
			}
		}
		for _, cmd := range commandsGui {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, cmd.ID)
			if err != nil {
				log.Printf("Cannot delete guild command %s: %v", cmd.Name, err)
			} else {
				log.Printf("Deleted guild command: %s", cmd.Name)
			}
		}

	}

	log.Println("Gracefully shutting down.")
}