package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

type Quest struct {
	ID            string
	Name          string
	Difficulty    Difficulty
	Target        int
	ScoreboardObj string
	ScoreboardCri string
	RewardCmd     string
	RewardText    string
}

type DailyState struct {
	ActiveQuests map[Difficulty]Quest
	ResetTime    time.Time
	Session      *discordgo.Session
	ChannelID    string
	Mutex        sync.Mutex
}

var CurrentDaily = &DailyState{
	ActiveQuests: make(map[Difficulty]Quest),
}

var allQuests = []Quest{
	// Easy
	{"E1", "Beri makan Hewan", Easy, 10, "daily_easy", "minecraft.custom:minecraft.animals_bred", "give %s cooked_beef 2", "2 Daging Matang"},
	{"E2", "Panen Gandum", Easy, 32, "daily_easy", "minecraft.mined:minecraft.wheat", "give %s bread 5", "5 Roti"},
	{"E3", "Tebang Kayu Oak", Easy, 50, "daily_easy", "minecraft.mined:minecraft.oak_log", "give %s iron_ingot 1", "1 Iron Ingot"},

	// Medium
	{"M1", "Bunuh Zombie", Medium, 15, "daily_med", "minecraft.killed:minecraft.zombie", "give %s emerald 1", "1 Emerald"},
	{"M2", "Tambang Iron Ore", Medium, 30, "daily_med", "minecraft.mined:minecraft.iron_ore", "give %s gold_ingot 3", "3 Gold Ingot"},
	{"M3", "Tangkap Ikan", Medium, 5, "daily_med", "minecraft.custom:minecraft.fish_caught", "experience add %s 100", "100 XP"},

	// Hard
	{"H1", "Cari Jamur Nether (Crimson)", Hard, 20, "daily_hard", "minecraft.mined:minecraft.crimson_fungus", "give %s diamond 1", "1 Diamond"},
	{"H2", "Bunuh Enderman", Hard, 5, "daily_hard", "minecraft.killed:minecraft.enderman", "give %s golden_apple 1", "1 Golden Apple"},
	{"H3", "Tambang Ancient Debris", Hard, 1, "daily_hard", "minecraft.mined:minecraft.ancient_debris", "give %s netherite_scrap 1", "1 Netherite Scrap"},
}

// InitDailySystem starts the daily quest ticker and initialization
func InitDailySystem(s *discordgo.Session, channelID string) {
	CurrentDaily.Session = s
	CurrentDaily.ChannelID = channelID

	// Random seed
	rand.Seed(time.Now().UnixNano())

	// Initialize quests on startup
	generateDailyQuests()

	go func() {
		// We use 30 second ticker to check and grant rewards fast
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			checkQuestProgress()

			CurrentDaily.Mutex.Lock()
			if time.Now().After(CurrentDaily.ResetTime) {
				CurrentDaily.Mutex.Unlock()
				generateDailyQuests()
				if CurrentDaily.Session != nil && CurrentDaily.ChannelID != "" {
					CurrentDaily.Session.ChannelMessageSend(CurrentDaily.ChannelID, "🔄 **Misi Harian Telah Di-Reset!**\nKetik `/daily` untuk melihat misi baru hari ini!")
				}
			} else {
				CurrentDaily.Mutex.Unlock()
			}
		}
	}()
}

func generateDailyQuests() {
	CurrentDaily.Mutex.Lock()
	defer CurrentDaily.Mutex.Unlock()

	// Clear previous scoreboards just in case
	runRcon("scoreboard objectives remove daily_easy")
	runRcon("scoreboard objectives remove daily_med")
	runRcon("scoreboard objectives remove daily_hard")

	var easy, med, hard []Quest
	for _, q := range allQuests {
		if q.Difficulty == Easy {
			easy = append(easy, q)
		} else if q.Difficulty == Medium {
			med = append(med, q)
		} else if q.Difficulty == Hard {
			hard = append(hard, q)
		}
	}

	eQuest := easy[rand.Intn(len(easy))]
	mQuest := med[rand.Intn(len(med))]
	hQuest := hard[rand.Intn(len(hard))]

	CurrentDaily.ActiveQuests[Easy] = eQuest
	CurrentDaily.ActiveQuests[Medium] = mQuest
	CurrentDaily.ActiveQuests[Hard] = hQuest

	// Set next reset to tomorrow 00:00:00
	now := time.Now()
	nextReset := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	// If it's already past 00:00, it sets to next day correctly
	CurrentDaily.ResetTime = nextReset

	// Add new scoreboards
	runRcon(fmt.Sprintf(`scoreboard objectives add daily_easy %s "%s"`, eQuest.ScoreboardCri, eQuest.Name))
	runRcon(fmt.Sprintf(`scoreboard objectives add daily_med %s "%s"`, mQuest.ScoreboardCri, mQuest.Name))
	runRcon(fmt.Sprintf(`scoreboard objectives add daily_hard %s "%s"`, hQuest.ScoreboardCri, hQuest.Name))

	// Display only one on sidebar (e.g. Hard mission)
	runRcon("scoreboard objectives setdisplay sidebar daily_hard")
}

func checkQuestProgress() {
	CurrentDaily.Mutex.Lock()
	quests := []Quest{
		CurrentDaily.ActiveQuests[Easy],
		CurrentDaily.ActiveQuests[Medium],
		CurrentDaily.ActiveQuests[Hard],
	}
	// s := CurrentDaily.Session
	// ch := CurrentDaily.ChannelID
	CurrentDaily.Mutex.Unlock()

	for _, q := range quests {
		// Use Minecraft selector to target everyone who reached the score!
		rewardCmd := strings.Replace(q.RewardCmd, "%s", "@s", -1)

		// 1. Give reward
		runRcon(fmt.Sprintf("execute as @a[scores={%s=%d..}] run %s", q.ScoreboardObj, q.Target, rewardCmd))

		// 2. Play victory sound
		runRcon(fmt.Sprintf("execute as @a[scores={%s=%d..}] run playsound entity.player.levelup master @s", q.ScoreboardObj, q.Target))
		
		// 3. Announce in-game
		runRcon(fmt.Sprintf(`execute as @a[scores={%s=%d..}] run tellraw @a [{"text":"[Daily Quest] ","color":"green"},{"selector":"@s"},{"text":" telah menyelesaikan misi ","color":"yellow"},{"text":"%s","color":"gold","bold":true},{"text":"!","color":"yellow"}]`, q.ScoreboardObj, q.Target, q.Name))

		// 4. Mark as complete by setting their score to negative so they don't get rewards repeatedly today
		runRcon(fmt.Sprintf("scoreboard players set @a[scores={%s=%d..}] %s -99999", q.ScoreboardObj, q.Target, q.ScoreboardObj))
	}
}

// GenerateDailyEmbed creates a rich discord embed for /daily
func GenerateDailyEmbed() *discordgo.MessageEmbed {
	CurrentDaily.Mutex.Lock()
	defer CurrentDaily.Mutex.Unlock()

	embed := &discordgo.MessageEmbed{
		Title: "📜 MISI HARIAN HARI INI",
		Color: 0x00FF00,
		Description: fmt.Sprintf("Misi akan di-reset pada: <t:%d:R>\n\nSelesaikan misi di bawah ini secara otomatis di dalam game untuk mendapatkan hadiah masuk ke dalam tas Anda!", CurrentDaily.ResetTime.Unix()),
	}

	if q, ok := CurrentDaily.ActiveQuests[Easy]; ok {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "🟢 Mudah",
			Value: fmt.Sprintf("**%s**\n🎯 Target: %d\n🎁 Hadiah: %s", q.Name, q.Target, q.RewardText),
		})
	}
	if q, ok := CurrentDaily.ActiveQuests[Medium]; ok {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "🟡 Sedang",
			Value: fmt.Sprintf("**%s**\n🎯 Target: %d\n🎁 Hadiah: %s", q.Name, q.Target, q.RewardText),
		})
	}
	if q, ok := CurrentDaily.ActiveQuests[Hard]; ok {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "🔴 Sulit",
			Value: fmt.Sprintf("**%s**\n🎯 Target: %d\n🎁 Hadiah: %s", q.Name, q.Target, q.RewardText),
		})
	}

	return embed
}
