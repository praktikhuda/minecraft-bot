package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func runRconWithReply(cmdStr string) (string, error) {
	rconPass := os.Getenv("RCON_PASSWORD")
	rconPort := os.Getenv("RCON_PORT")
	if rconPort == "" {
		rconPort = "25575"
	}
	cmd := exec.Command("mcrcon", "-H", "127.0.0.1", "-P", rconPort, "-p", rconPass, cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("RCON failed for %s: %v", cmdStr, err)
		return "", err
	}
	return string(out), nil
}

// Pos3D represents 3D coordinates
type Pos3D struct {
	X float64
	Y float64
	Z float64
}

// GetPlayerPos uses RCON to get a player's coordinates
func GetPlayerPos(playerName string) (Pos3D, error) {
	out, err := runRconWithReply(fmt.Sprintf("data get entity %s Pos", playerName))
	if err != nil {
		return Pos3D{}, err
	}

	re := regexp.MustCompile(`\[([^d]+)d, ([^d]+)d, ([^d]+)d\]`)
	matches := re.FindStringSubmatch(out)
	if len(matches) != 4 {
		return Pos3D{}, fmt.Errorf("failed to parse position from: %s", out)
	}

	x, _ := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64)
	y, _ := strconv.ParseFloat(strings.TrimSpace(matches[2]), 64)
	z, _ := strconv.ParseFloat(strings.TrimSpace(matches[3]), 64)

	return Pos3D{X: x, Y: y, Z: z}, nil
}

// GetPlayerDimension gets the dimension of a player
func GetPlayerDimension(playerName string) (string, error) {
	out, err := runRconWithReply(fmt.Sprintf("data get entity %s Dimension", playerName))
	if err != nil {
		return "", err
	}

	if strings.Contains(out, "minecraft:the_nether") {
		return "minecraft:the_nether", nil
	} else if strings.Contains(out, "minecraft:the_end") {
		return "minecraft:the_end", nil
	}
	return "minecraft:overworld", nil
}

// StoreInventoryInVault moves player inventory to the vault chest block
func StoreInventoryInVault(playerName string, vaultX, vaultY, vaultZ int) {
	slots := []string{}
	for i := 0; i <= 26; i++ {
		slots = append(slots, fmt.Sprintf("inventory.%d", i))
	}
	for i := 0; i <= 8; i++ {
		slots = append(slots, fmt.Sprintf("hotbar.%d", i))
	}
	slots = append(slots, "armor.head", "armor.chest", "armor.legs", "armor.feet", "weapon.offhand")

	for i, slotName := range slots {
		y := vaultY
		slotIdx := i
		if i > 26 {
			y = vaultY + 1
			slotIdx = i - 27
		}
		runRcon(fmt.Sprintf("item replace block %d %d %d container.%d from entity %s %s", vaultX, y, vaultZ, slotIdx, playerName, slotName))
		time.Sleep(50 * time.Millisecond) // avoid spamming server too hard
	}
}

// RestoreInventoryFromVault moves items back from the vault to the player
func RestoreInventoryFromVault(playerName string, vaultX, vaultY, vaultZ int) {
	slots := []string{}
	for i := 0; i <= 26; i++ {
		slots = append(slots, fmt.Sprintf("inventory.%d", i))
	}
	for i := 0; i <= 8; i++ {
		slots = append(slots, fmt.Sprintf("hotbar.%d", i))
	}
	slots = append(slots, "armor.head", "armor.chest", "armor.legs", "armor.feet", "weapon.offhand")

	for i, slotName := range slots {
		y := vaultY
		slotIdx := i
		if i > 26 {
			y = vaultY + 1
			slotIdx = i - 27
		}
		runRcon(fmt.Sprintf("item replace entity %s %s from block %d %d %d container.%d", playerName, slotName, vaultX, y, vaultZ, slotIdx))
		time.Sleep(50 * time.Millisecond)
	}

	for i := 0; i < len(slots); i++ {
		y := vaultY
		slotIdx := i
		if i > 26 {
			y = vaultY + 1
			slotIdx = i - 27
		}
		runRcon(fmt.Sprintf("item replace block %d %d %d container.%d with air", vaultX, y, vaultZ, slotIdx))
	}
}
