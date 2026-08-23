package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type TourneyState struct {
	IsActive   bool
	Player1    string
	Player2    string
	P1Pos      Pos3D
	P1Dim      string
	P2Pos      Pos3D
	P2Dim      string
	Vault1X    int
	Vault1Y    int
	Vault1Z    int
	Vault2X    int
	Vault2Y    int
	Vault2Z    int
	Timer      *time.Timer
	Mutex      sync.Mutex
	ChannelID  string
	Session    *discordgo.Session
}

var CurrentTourney = &TourneyState{}

// Config default
var (
	ArenaX = 1000
	ArenaY = 100
	ArenaZ = 1000
	
	// Vault coordinates (we use 2 separate double chests for 2 players)
	BaseVaultX = 0
	BaseVaultY = -60
	BaseVaultZ = 0
)

// PrepareChest ensures a double chest exists at the vault coordinates
func PrepareChest(x, y, z int) {
	// A double chest requires 2 chests side by side
	runRcon(fmt.Sprintf("setblock %d %d %d chest", x, y, z))
	runRcon(fmt.Sprintf("setblock %d %d %d chest", x+1, y, z))
}

// GiveStarterPack gives the gladiator kit
func GiveStarterPack(player string) {
	runRcon(fmt.Sprintf("clear %s", player))
	time.Sleep(100 * time.Millisecond)
	
	runRcon(fmt.Sprintf("item replace entity %s armor.head with iron_helmet", player))
	runRcon(fmt.Sprintf("item replace entity %s armor.chest with iron_chestplate", player))
	runRcon(fmt.Sprintf("item replace entity %s armor.legs with iron_leggings", player))
	runRcon(fmt.Sprintf("item replace entity %s armor.feet with iron_boots", player))
	runRcon(fmt.Sprintf("item replace entity %s weapon.offhand with shield", player))
	
	runRcon(fmt.Sprintf("give %s iron_sword 1", player))
	runRcon(fmt.Sprintf("give %s cooked_beef 64", player))
	runRcon(fmt.Sprintf("give %s golden_apple 1", player))
	
	runRcon(fmt.Sprintf("effect give %s instant_health 1 10", player))
	runRcon(fmt.Sprintf("effect give %s saturation 1 10", player))
}

func StartTourney(s *discordgo.Session, i *discordgo.InteractionCreate, p1, p2 string) {
	CurrentTourney.Mutex.Lock()
	if CurrentTourney.IsActive {
		CurrentTourney.Mutex.Unlock()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ Sedang ada turnamen yang berlangsung! Tunggu sampai selesai."},
		})
		return
	}
	
	CurrentTourney.IsActive = true
	CurrentTourney.Player1 = p1
	CurrentTourney.Player2 = p2
	CurrentTourney.ChannelID = i.ChannelID
	CurrentTourney.Session = s
	
	CurrentTourney.Vault1X = BaseVaultX
	CurrentTourney.Vault1Y = BaseVaultY
	CurrentTourney.Vault1Z = BaseVaultZ
	
	CurrentTourney.Vault2X = BaseVaultX + 2
	CurrentTourney.Vault2Y = BaseVaultY
	CurrentTourney.Vault2Z = BaseVaultZ

	CurrentTourney.Mutex.Unlock()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: fmt.Sprintf("⚔️ **TURNAMEN DIMULAI!**\nGladiator: **%s** VS **%s**\nMempersiapkan arena dan menculik pemain...", p1, p2)},
	})

	go func() {
		// 1. Get positions
		p1pos, err1 := GetPlayerPos(p1)
		p1dim, _ := GetPlayerDimension(p1)
		p2pos, err2 := GetPlayerPos(p2)
		p2dim, _ := GetPlayerDimension(p2)

		if err1 != nil || err2 != nil {
			s.ChannelMessageSend(i.ChannelID, "❌ Gagal mendeteksi lokasi pemain. Pastikan keduanya sedang online!")
			ResetTourney()
			return
		}

		CurrentTourney.P1Pos = p1pos
		CurrentTourney.P1Dim = p1dim
		CurrentTourney.P2Pos = p2pos
		CurrentTourney.P2Dim = p2dim

		// 2. Prepare Vault Chests
		PrepareChest(CurrentTourney.Vault1X, CurrentTourney.Vault1Y, CurrentTourney.Vault1Z)
		PrepareChest(CurrentTourney.Vault2X, CurrentTourney.Vault2Y, CurrentTourney.Vault2Z)

		// 3. Store Inventories
		s.ChannelMessageSend(i.ChannelID, "📦 Menyimpan barang bawaan peserta ke dalam brankas...")
		StoreInventoryInVault(p1, CurrentTourney.Vault1X, CurrentTourney.Vault1Y, CurrentTourney.Vault1Z)
		StoreInventoryInVault(p2, CurrentTourney.Vault2X, CurrentTourney.Vault2Y, CurrentTourney.Vault2Z)

		// 4. Teleport to Arena & Gear Up
		runRcon(fmt.Sprintf("execute in minecraft:overworld run tp %s %d %d %d", p1, ArenaX-5, ArenaY, ArenaZ))
		runRcon(fmt.Sprintf("execute in minecraft:overworld run tp %s %d %d %d", p2, ArenaX+5, ArenaY, ArenaZ))

		GiveStarterPack(p1)
		GiveStarterPack(p2)

		// 5. Title & Fight
		runRcon(fmt.Sprintf("title %s title {\"text\":\"FIGHT!\",\"color\":\"red\",\"bold\":true}", p1))
		runRcon(fmt.Sprintf("title %s title {\"text\":\"FIGHT!\",\"color\":\"red\",\"bold\":true}", p2))

		// 6. Set 10 Minute Timer
		CurrentTourney.Timer = time.AfterFunc(10*time.Minute, func() {
			EndTourney("DRAW", "")
		})
	}()
}

func restorePlayer(p string, pos Pos3D, dim string, vaultX, vaultY, vaultZ int) {
	go func() {
		timeout := time.Now().Add(5 * time.Minute)
		for time.Now().Before(timeout) {
			out, _ := runRconWithReply(fmt.Sprintf("data get entity %s Health", p))
			if !strings.Contains(out, "No entity was found") && !strings.Contains(out, " 0.0f") && !strings.Contains(out, " 0f") {
				time.Sleep(1 * time.Second)
				runRcon(fmt.Sprintf("execute in %s run tp %s %f %f %f", dim, p, pos.X, pos.Y, pos.Z))
				time.Sleep(1 * time.Second)
				runRcon(fmt.Sprintf("clear %s", p))
				RestoreInventoryFromVault(p, vaultX, vaultY, vaultZ)
				return
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func EndTourney(reason, winner string) {
	CurrentTourney.Mutex.Lock()
	if !CurrentTourney.IsActive {
		CurrentTourney.Mutex.Unlock()
		return
	}
	
	if CurrentTourney.Timer != nil {
		CurrentTourney.Timer.Stop()
	}
	
	p1 := CurrentTourney.Player1
	p2 := CurrentTourney.Player2
	CurrentTourney.IsActive = false
	s := CurrentTourney.Session
	ch := CurrentTourney.ChannelID
	CurrentTourney.Mutex.Unlock()

	if s != nil && ch != "" {
		if reason == "DRAW" {
			s.ChannelMessageSend(ch, fmt.Sprintf("⏳ **WAKTU HABIS (10 Menit)!**\nPertarungan **%s** vs **%s** berakhir SERI! Mengembalikan pemain ke asal...", p1, p2))
		} else if reason == "KILL" {
			s.ChannelMessageSend(ch, fmt.Sprintf("🏆 **PERTARUNGAN SELESAI!**\nPemenangnya adalah: **%s**!\nMengembalikan pemain ke asal...", winner))
		}
	}

	// Bersihkan dropped items di dalam arena
	go func() {
		runRcon(fmt.Sprintf("kill @e[type=item,x=%d,y=%d,z=%d,dx=30,dy=7,dz=30]", ArenaX-15, ArenaY-1, ArenaZ-15))
	}()

	restorePlayer(p1, CurrentTourney.P1Pos, CurrentTourney.P1Dim, CurrentTourney.Vault1X, CurrentTourney.Vault1Y, CurrentTourney.Vault1Z)
	restorePlayer(p2, CurrentTourney.P2Pos, CurrentTourney.P2Dim, CurrentTourney.Vault2X, CurrentTourney.Vault2Y, CurrentTourney.Vault2Z)
}

func ResetTourney() {
	CurrentTourney.Mutex.Lock()
	CurrentTourney.IsActive = false
	if CurrentTourney.Timer != nil {
		CurrentTourney.Timer.Stop()
	}
	CurrentTourney.Mutex.Unlock()
}

// HandleKillEvent should be called by LogListen when a death is detected
func HandleKillEvent(logLine string) {
	CurrentTourney.Mutex.Lock()
	if !CurrentTourney.IsActive {
		CurrentTourney.Mutex.Unlock()
		return
	}
	
	p1 := CurrentTourney.Player1
	p2 := CurrentTourney.Player2
	CurrentTourney.Mutex.Unlock()

	// Typical death log: "Devan was slain by MrPheee" or "MrPheee fell from a high place"
	// We just check if p1 or p2 name is in the line and someone died.
	// Since it's an arena, if p1 dies, p2 wins.
	
	if strings.Contains(logLine, p1) && strings.Contains(logLine, "was ") { // p1 died
		EndTourney("KILL", p2)
		return
	}
	if strings.Contains(logLine, p2) && strings.Contains(logLine, "was ") { // p2 died
		EndTourney("KILL", p1)
		return
	}
	
	// Fallback for non 'was slain by' (e.g. burned to death, fell)
	if strings.Contains(logLine, p1+" died") || strings.Contains(logLine, p1+" fell") || strings.Contains(logLine, p1+" burned") || strings.Contains(logLine, p1+" hit the ground") || strings.Contains(logLine, p1+" went up in flames") || strings.Contains(logLine, p1+" suffocated") || strings.Contains(logLine, p1+" starved") {
		EndTourney("KILL", p2)
		return
	}
	if strings.Contains(logLine, p2+" died") || strings.Contains(logLine, p2+" fell") || strings.Contains(logLine, p2+" burned") || strings.Contains(logLine, p2+" hit the ground") || strings.Contains(logLine, p2+" went up in flames") || strings.Contains(logLine, p2+" suffocated") || strings.Contains(logLine, p2+" starved") {
		EndTourney("KILL", p1)
		return
	}
}
