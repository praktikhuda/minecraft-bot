package utils

import (
	"fmt"
		"strings"
)

// SpawnBlackMarket generates random coordinates and spawns a custom wandering trader.
func SpawnBlackMarket() (string, error) {
	// Cek apakah pedagang gelap sudah ada di dunia
	output, err := sendRconCommand("execute if entity @e[type=wandering_trader,name=\"Pedagang Gelap\"]")
	if err == nil && strings.Contains(output, "Test passed") {
		return "❌ **PANGGILAN GAGAL:** Pedagang Gelap saat ini masih bersembunyi di suatu tempat! Temukan dia atau tunggu sampai dia menghilang (30 Menit) sebelum memanggil yang baru.", nil
	}

	// Initialize random seed (disabled for developer phase)

	// FASE DEVELOPER: Koordinat statis sesuai permintaan
	x := -10
	y := 75
	z := 5

	// RCON command to spawn the wandering trader
	summonCmd := fmt.Sprintf(
		`execute in minecraft:overworld positioned %d %d %d run summon wandering_trader ~ ~ ~ {CustomName:'"Pedagang Gelap"',Invulnerable:1b,DespawnDelay:36000,Attributes:[{id:"minecraft:generic.movement_speed",base:0.0}],Offers:{Recipes:[{buy:{id:"minecraft:wither_skeleton_skull",count:1},buyB:{id:"minecraft:diamond",count:64},sell:{id:"minecraft:elytra",count:1},maxUses:1,rewardExp:0b}]}}`,
		x, y, z,
	)

	// Broadcast to the server
	broadcastCmd := fmt.Sprintf(
		`tellraw @a {"text":"\n[🚨 EVENT BLACK MARKET 🚨]\nPedagang Gelap telah muncul di koordinat X: %d, Z: %d!\nDia menjual Elytra seharga 1 Wither Skeleton Skull dan 64 Diamond. Stok hanya 1! Siapa cepat dia dapat!\n","color":"yellow","bold":true}`,
		x, z,
	)

	// Execute RCON
	err = runRcon(summonCmd)
	if err != nil {
		return "", fmt.Errorf("gagal memanggil pedagang: %w", err)
	}

	err = runRcon(broadcastCmd)
	if err != nil {
		// Log the error but don't fail the command if broadcasting fails
		fmt.Printf("Gagal broadcast blackmarket: %v\n", err)
	}

	// Message to return to Discord
	discordMsg := fmt.Sprintf("**🚨 EVENT BLACK MARKET DIMULAI 🚨**\n\nPedagang Gelap telah dipanggil di koordinat **X: %d, Z: %d**!\n\nBarang yang dijual:\n- 🪽 **1x Elytra**\nHarga:\n- 💀 **1x Wither Skeleton Skull**\n- 💎 **64x Diamond**\n\nStok hanya **1**. Waktu memburu: **30 Menit**!", x, z)
	
	return discordMsg, nil
}

// RemoveBlackMarket kills all wandering traders in the world.
func RemoveBlackMarket() (string, error) {
	// Execute kill command for all wandering traders to clean up the bugged ones
	killCmd := `execute in minecraft:overworld run kill @e[type=wandering_trader]`
	err := runRcon(killCmd)
	if err != nil {
		return "❌ Gagal mengusir pedagang: " + err.Error(), err
	}
	return "🧹 **PEMBERSIHAN BERHASIL!**\nSeluruh Pedagang Gelap (termasuk yang bug) telah diusir paksa dari dunia!", nil
}
