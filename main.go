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
			Name:                     "sync_titles",
			Description:              "Update ingame titles for top players automatically",
			DefaultMemberPermissions: &defaultMemberPermisions,
		},
		{
			Name:        "leaderboard",
			Description: "Lihat papan peringkat",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "combat",
					Description: "Statistik Pertarungan PVP & PVE",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "kategori",
							Description: "Pilih kategori",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Raja PVP",
									Value: "minecraft:custom:player_kills",
								},
								{
									Name:  "Ahli Panah",
									Value: "minecraft:used:bow",
								},
								{
									Name:  "Sniper Jarak Jauh",
									Value: "minecraft:custom:target_hit",
								},
								{
									Name:  "Penguasa Trident",
									Value: "minecraft:used:trident",
								},
								{
									Name:  "Sembilan Nyawa",
									Value: "minecraft:used:totem_of_undying",
								},
								{
									Name:  "Si Paling Samsak",
									Value: "minecraft:custom:damage_taken",
								},
								{
									Name:  "Haus Darah",
									Value: "minecraft:custom:damage_dealt",
								},
								{
									Name:  "Ahli Pedang",
									Value: "minecraft:used:netherite_sword",
								},
								{
									Name:  "Si Paling Rapuh",
									Value: "minecraft:custom:damage_resisted",
								},

							},
						},
					},
				},
				{
					Name:        "hunt",
					Description: "Statistik Membunuh Monster & Boss",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "kategori",
							Description: "Pilih kategori",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Pembantai Zombie",
									Value: "minecraft:killed:zombie",
								},
								{
									Name:  "Pembantai Skeleton",
									Value: "minecraft:killed:skeleton",
								},
								{
									Name:  "Pembantai Creeper",
									Value: "minecraft:killed:creeper",
								},
								{
									Name:  "Pembunuh Naga",
									Value: "minecraft:killed:ender_dragon",
								},
								{
									Name:  "Penjaga Desa",
									Value: "minecraft:killed:pillager",
								},
								{
									Name:  "Gladiator",
									Value: "minecraft:killed:wither",
								},
								{
									Name:  "Pemburu Phantom",
									Value: "minecraft:killed:phantom",
								},
								{
									Name:  "Tukang Tipu Piglin",
									Value: "minecraft:killed:piglin",
								},
								{
									Name:  "Pemburu Enderman",
									Value: "minecraft:killed:enderman",
								},

							},
						},
					},
				},
				{
					Name:        "mining",
					Description: "Statistik Menambang & Pekerja",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "kategori",
							Description: "Pilih kategori",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Juragan Diamond",
									Value: "minecraft:mined:diamond_ore",
								},
								{
									Name:  "Pencari Netherite",
									Value: "minecraft:mined:ancient_debris",
								},
								{
									Name:  "Kuli Batu",
									Value: "minecraft:mined:stone",
								},
								{
									Name:  "Tukang Gali",
									Value: "minecraft:mined:dirt",
								},
								{
									Name:  "Raja Hutan",
									Value: "minecraft:mined:oak_log",
								},
								{
									Name:  "Penambang Emas",
									Value: "minecraft:mined:gold_ore",
								},
								{
									Name:  "Petani Gandum",
									Value: "minecraft:mined:wheat",
								},
								{
									Name:  "Petani Wortel",
									Value: "minecraft:mined:carrots",
								},
								{
									Name:  "Pemecah Pickaxe",
									Value: "minecraft:broken:diamond_pickaxe",
								},
								{
									Name:  "Penggila Redstone",
									Value: "minecraft:mined:redstone_ore",
								},
								{
									Name:  "Tukang Gali Pasir",
									Value: "minecraft:mined:sand",
								},
								{
									Name:  "Penambang Lapis",
									Value: "minecraft:mined:lapis_ore",
								},
								{
									Name:  "Pembersih Nether",
									Value: "minecraft:mined:netherrack",
								},

							},
						},
					},
				},
				{
					Name:        "explore",
					Description: "Statistik Penjelajah & Waktu",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "kategori",
							Description: "Pilih kategori",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Tukang Kabur",
									Value: "minecraft:custom:leave_game",
								},
								{
									Name:  "Si Paling Panik",
									Value: "minecraft:custom:crouch_one_cm",
								},
								{
									Name:  "Phobia Ketinggian",
									Value: "minecraft:custom:climb_one_cm",
								},
								{
									Name:  "Hobi Cuci Muka",
									Value: "minecraft:custom:walk_under_water_one_cm",
								},
								{
									Name:  "Si Paling Sibuk",
									Value: "minecraft:custom:walk_one_cm",
								},
								{
									Name:  "Pelari Cepat",
									Value: "minecraft:custom:sprint_one_cm",
								},
								{
									Name:  "Pengelana Air",
									Value: "minecraft:custom:boat_one_cm",
								},
								{
									Name:  "Penunggang Babi",
									Value: "minecraft:custom:pig_one_cm",
								},
								{
									Name:  "Joki Kuda",
									Value: "minecraft:custom:horse_one_cm",
								},
								{
									Name:  "Pilot Handal",
									Value: "minecraft:custom:aviate_one_cm",
								},
								{
									Name:  "Perenang Cepat",
									Value: "minecraft:custom:swim_one_cm",
								},
								{
									Name:  "Tukang Lompat Es",
									Value: "minecraft:custom:walk_on_water_one_cm",
								},
								{
									Name:  "Pengembara Nether",
									Value: "minecraft:custom:strider_one_cm",
								},
								{
									Name:  "Tukang Nyasar",
									Value: "minecraft:custom:play_time",
								},
								{
									Name:  "Turis Dimensi",
									Value: "minecraft:custom:time_since_last_rest",
								},

							},
						},
					},
				},
				{
					Name:        "food",
					Description: "Statistik Makanan & Ramuan",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "kategori",
							Description: "Pilih kategori",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Makan Sembarangan",
									Value: "minecraft:used:rotten_flesh",
								},
								{
									Name:  "Tukang Makan",
									Value: "minecraft:custom:eat_record_count",
								},
								{
									Name:  "Pemakan Apel Emas",
									Value: "minecraft:used:golden_apple",
								},
								{
									Name:  "Pemakan Apel Enchant",
									Value: "minecraft:used:enchanted_golden_apple",
								},
								{
									Name:  "Tukang Roti",
									Value: "minecraft:used:bread",
								},
								{
									Name:  "Pemabuk Potion",
									Value: "minecraft:used:potion",
								},
								{
									Name:  "Pecandu Madu",
									Value: "minecraft:used:honey_bottle",
								},

							},
						},
					},
				},
				{
					Name:        "fails",
					Description: "Statistik Kesialan & Kematian",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "kategori",
							Description: "Pilih kategori",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Korban Ledakan",
									Value: "minecraft:killed_by:creeper",
								},
								{
									Name:  "Jatuh Dari Langit",
									Value: "minecraft:custom:fall_one_cm",
								},
								{
									Name:  "Mandi Lava",
									Value: "minecraft:killed_by:lava",
								},
								{
									Name:  "Tersambar Petir",
									Value: "minecraft:killed_by:lightning_bolt",
								},
								{
									Name:  "Dilempar Kinetik",
									Value: "minecraft:killed_by:fly_into_wall",
								},
								{
									Name:  "Korban Llama",
									Value: "minecraft:killed_by:llama_spit",
								},
								{
									Name:  "Si Paling Tumbal",
									Value: "minecraft:custom:deaths",
								},
								{
									Name:  "Korban Gravitasi",
									Value: "minecraft:killed_by:fall",
								},
								{
									Name:  "Lemah Jantung",
									Value: "minecraft:killed_by:wither_skeleton",
								},
								{
									Name:  "Korban Kaktus",
									Value: "minecraft:killed_by:cactus",
								},
								{
									Name:  "Mati Konyol",
									Value: "minecraft:killed_by:cramming",
								},
								{
									Name:  "Mati Tertimpa Anvil",
									Value: "minecraft:killed_by:falling_block",
								},
								{
									Name:  "Si Buta",
									Value: "minecraft:custom:time_since_death",
								},

							},
						},
					},
				},
				{
					Name:        "misc",
					Description: "Statistik Ekonomi & Hobi",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "kategori",
							Description: "Pilih kategori",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Si Paling Capek",
									Value: "minecraft:custom:jump",
								},
								{
									Name:  "Kaum Rebahan",
									Value: "minecraft:custom:sleep_in_bed",
								},
								{
									Name:  "Cinta Damai",
									Value: "minecraft:custom:animals_bred",
								},
								{
									Name:  "Tukang Mancing",
									Value: "minecraft:custom:fish_caught",
								},
								{
									Name:  "Pencari Harta",
									Value: "minecraft:custom:inspect_hopper",
								},
								{
									Name:  "Juragan Kampung",
									Value: "minecraft:custom:traded_with_villager",
								},
								{
									Name:  "Pembakar Makanan",
									Value: "minecraft:used:furnace",
								},
								{
									Name:  "Pencari XP",
									Value: "minecraft:custom:total_world_time",
								},
								{
									Name:  "Si Paling Penempatan",
									Value: "minecraft:used:cobblestone",
								},
								{
									Name:  "Desainer Interior",
									Value: "minecraft:used:oak_planks",
								},
								{
									Name:  "Tukang Kaca",
									Value: "minecraft:used:glass",
								},
								{
									Name:  "Pembakar Obor",
									Value: "minecraft:used:torch",
								},
								{
									Name:  "Ahli Beton",
									Value: "minecraft:used:cyan_concrete",
								},
								{
									Name:  "Pembuat Pintu",
									Value: "minecraft:used:iron_door",
								},
								{
									Name:  "Tukang Ledak",
									Value: "minecraft:used:tnt",
								},
								{
									Name:  "Pemasang Kasur",
									Value: "minecraft:used:red_bed",
								},
								{
									Name:  "Kolektor Musik",
									Value: "minecraft:used:music_disc_13",
								},
								{
									Name:  "Petasan Mania",
									Value: "minecraft:used:firework_rocket",
								},
								{
									Name:  "Ahli Lonceng",
									Value: "minecraft:custom:bell_ring",
								},
								{
									Name:  "Seniman Banner",
									Value: "minecraft:used:loom",
								},

							},
						},
					},
				},
				{
					Name:        "extra",
					Description: "Statistik Lain-Lain",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "kategori",
							Description: "Pilih kategori",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Pembuat Peta",
									Value: "minecraft:used:cartography_table",
								},
								{
									Name:  "Pandai Besi",
									Value: "minecraft:used:anvil",
								},
								{
									Name:  "Tukang Enchanter",
									Value: "minecraft:used:enchanting_table",
								},
								{
									Name:  "Penebar Bunga",
									Value: "minecraft:used:poppy",
								},
								{
									Name:  "Pewarna Domba",
									Value: "minecraft:used:red_dye",
								},
								{
									Name:  "Pengepul Shulker",
									Value: "minecraft:used:shulker_box",
								},
								{
									Name:  "Tukang Nge-Camp",
									Value: "minecraft:used:campfire",
								},
								{
									Name:  "Raja Sampah",
									Value: "minecraft:dropped:cobblestone",
								},
								{
									Name:  "Tukang Bersih-Bersih",
									Value: "minecraft:picked_up:cobblestone",
								},
								{
									Name:  "Penggila Telur",
									Value: "minecraft:used:egg",
								},
								{
									Name:  "Pengendali Salju",
									Value: "minecraft:used:snowball",
								},
								{
									Name:  "Penjaga Ender Chest",
									Value: "minecraft:used:ender_chest",
								},

							},
						},
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
			Name:        "wllist",
			Description: "Melihat daftar pemain yang ada di whitelist",
		},
		{
			Name:        "tourney",
			Description: "Mulai turnamen PVP 1-Lawan-1 di Arena",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player1",
					Description: "Nama pemain pertama di dalam game",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player2",
					Description: "Nama pemain kedua di dalam game",
					Required:    true,
				},
			},
		},
		{
			Name:        "gelar",
			Description: "Pilih gelar (title) mana yang ingin Anda pakai di dalam game",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "Username Minecraft Anda",
					Required:    true,
				},
			},
		},
		{
			Name:        "whereis",
			Description: "Melacak lokasi pemain (Admin Only)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "Nama pemain yang ingin dilacak (Kosongkan untuk cari semua)",
					Required:    false,
				},
			},
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
							Name:  "Spectator (Spectator + OP)",
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
			msg := "⚠️ **Sync Gelar dinonaktifkan** karena gelar sekarang dinamis."
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			if len(options) == 0 {
				return
			}
			
			subcommand := options[0]
			category := ""
			for _, opt := range subcommand.Options {
				if opt.Name == "kategori" {
					category = opt.StringValue()
				}
			}
			
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "📊 Sedang memuat leaderboard...",
				},
			})
			
			utils.MessageHandler("leaderboard", s, i, category)
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
		"listleaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			var kategori string
			if len(options) > 0 {
				kategori = options[0].StringValue()
			}
			utils.MessageHandler("listleaderboard", s, i, kategori)
		},
		"gelar": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			var username string
			if opt, ok := optionMap["username"]; ok {
				username = opt.StringValue()
			}

			// We won't respond here immediately because we might need to send a complex UI response
			// Wait, interaction must be responded to within 3 seconds. MessageHandler can do it.
			utils.MessageHandler("gelar", s, i, username)
		},
		"whereis": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**🌍 Melacak lokasi...**",
				},
			})
			options := i.ApplicationCommandData().Options
			username := ""
			for _, opt := range options {
				if opt.Name == "username" {
					username = opt.StringValue()
				}
			}
			utils.MessageHandler("whereis", s, i, username)
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

var componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"select_title": utils.HandleGelarSelection,
}

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
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
