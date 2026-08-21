package main

import (
	"flag"
	"fmt"
	"log"
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
	defaultMemberPermisions int64 = discordgo.PermissionAdministrator

	commands = []*discordgo.ApplicationCommand{
		{
			Name:                     "unban",
			Description:              "Unban a player from the server",
			DefaultMemberPermissions: &defaultMemberPermisions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "Player name to unban",
					Required:    true,
				},
			},
		},
		{
			Name:                     "rmblackmarket",
			Description:              "Hapus semua pedagang gelap dari dunia",
			DefaultMemberPermissions: &defaultMemberPermisions,
		},
		{
			Name:                     "blackmarket",
			Description:              "Panggil pedagang gelap ke koordinat acak",
			DefaultMemberPermissions: &defaultMemberPermisions,
		},
		{
			Name:                     "sync_titles",
			Description:              "Update ingame titles for top players automatically",
			DefaultMemberPermissions: &defaultMemberPermisions,
		},
		{
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
						{Name: "Sembilan Nyawa (Totem)", Value: "totems"},
					},
				},
			},
		},
		{
			Name:        "start",
			Description: "Start Minecraft Server",
		},
		{
			Name:                     "stop",
			Description:              "Stop Minecraft Server",
			DefaultMemberPermissions: &defaultMemberPermisions,
		},
		{
			Name:                     "restart",
			Description:              "Restart Minecraft Server",
			DefaultMemberPermissions: &defaultMemberPermisions,
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
			Name:                     "backup",
			Description:              "Create a ZIP backup of the Minecraft world (Admin Only)",
			DefaultMemberPermissions: &defaultMemberPermisions,
		},
		{
			Name:                     "logs",
			Description:              "Fetch the latest 15 lines of server logs (Admin Only)",
			DefaultMemberPermissions: &defaultMemberPermisions,
		},
		{
			Name:        "wllist",
			Description: "Melihat daftar pemain yang ada di whitelist",
		},
		{
			Name:        "whereis",
			Description: "Melacak lokasi seluruh pemain (Admin Only)",
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
			Name:                     "kick",
			Description:              "Kick a player from the Minecraft server",
			DefaultMemberPermissions: &defaultMemberPermisions,
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
			Name:                     "ban",
			Description:              "Ban a player from the Minecraft server",
			DefaultMemberPermissions: &defaultMemberPermisions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "The Minecraft username to ban",
					Required:    true,
				},
			},
		},
		{
			Name:                     "op",
			Description:              "Berikan jabatan khusus ke pemain Minecraft",
			DefaultMemberPermissions: &defaultMemberPermisions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "Username Minecraft",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "Pilih jabatan",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Admin (Creative + OP)",
							Value: "admin",
						},
						{
							Name:  "Spectator (Spectator + Deop)",
							Value: "spectator",
						},
						{
							Name:  "Player (Survival + Deop)",
							Value: "player",
						},
					},
				},
			},
		},
		{
			Name:                     "deop",
			Description:              "Remove Operator privileges from a player (Admin Only)",
			DefaultMemberPermissions: &defaultMemberPermisions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "The Minecraft username to De-OP",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"unban": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⏳ Membuka blokir pemain...",
				},
			})
			p := i.ApplicationCommandData().Options[0].StringValue()
			utils.MessageHandler("unban", s, i, p)
		},
		"rmblackmarket": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⏳ Sedang mengusir Pedagang Gelap...",
				},
			})
			msg, _ := utils.RemoveBlackMarket()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		"blackmarket": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⏳ Memanggil pedagang gelap dari langit...",
				},
			})
			msg, err := utils.SpawnBlackMarket()
			if err != nil {
				msg = "Error: " + err.Error()
			}
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		"sync_titles": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⏳ Sedang menyinkronkan gelar pemain...",
				},
			})
			msg, err := utils.SyncTitles()
			if err != nil {
				msg = "Error: " + err.Error()
			}
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
		},
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
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

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
		"tps": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Checking server TPS...**",
				},
			})
			utils.MessageHandler("tps", s, i, "")
		},
		"ip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Fetching Server IP...**",
				},
			})
			utils.MessageHandler("ip", s, i, "")
		},
		"backup": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Starting World Backup... Please wait.**",
				},
			})
			utils.MessageHandler("backup", s, i, "")
		},
		"logs": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Fetching the latest server logs...**",
				},
			})
			utils.MessageHandler("logs", s, i, "")
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
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

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
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

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
		"op": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

			options := i.ApplicationCommandData().Options
			username := ""
			role := ""
			for _, opt := range options {
				if opt.Name == "username" {
					username = opt.StringValue()
				} else if opt.Name == "role" {
					role = opt.StringValue()
				}
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("**Memproses jabatan `%s`...**", username),
				},
			})
			payload := username + ":" + role
			utils.MessageHandler("op", s, i, payload)
		},
		"deop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !isOwner(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ **Akses Ditolak:** Perintah ini dikunci dan hanya Owner Bot yang bisa memakainya!",
					},
				})
				return
			}

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
					Content: fmt.Sprintf("**De-opping `%s`...**", username),
				},
			})
			utils.MessageHandler("deop", s, i, username)
		},
		"whereis": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**🌍 Melacak lokasi seluruh pemain...**",
				},
			})
			utils.MessageHandler("whereis", s, i, "")
		},
		"wllist": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Mengambil data whitelist...**",
				},
			})
			utils.MessageHandler("wllist", s, i, "")
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


func isOwner(i *discordgo.InteractionCreate) bool {
	var userID string
	if i.Member != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}
	ownerID := os.Getenv("OWNER_ID")
	return userID == ownerID
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	s.AddHandler(utils.CrossChatHandler)

	s.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentMessageContent

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	utils.AutoStartLog(s)

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
