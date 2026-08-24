package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
)

func RunSetup() {
	if err := godotenv.Load("/home/samsul/script/go/mcbot/.env"); err != nil {
		log.Println("No .env found, skipping")
	}

	fmt.Println("Running RCON PAPI Commands...")
	runRcon("papi ecloud download Statistic")
	runRcon("papi ecloud download Objective")
	time.Sleep(2 * time.Second)
	runRcon("papi reload")
	fmt.Println("PAPI commands completed!")

	fmt.Println("Connecting to SSH to update scoreboards.yml...")
	config := &ssh.ClientConfig{
		User: "mcserv",
		Auth: []ssh.AuthMethod{
			ssh.Password("mc12345"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", "147.93.159.183:22", config)
	if err != nil {
		log.Fatalf("Failed to dial SSH: %s", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %s", err)
	}
	defer session.Close()

	yamlContent := `scoreboards:
  quests:
    update: 20
    conditions: []
    titles:
      - '&6&l====== DAILY QUESTS ======'
      - '&e&l====== DAILY QUESTS ======'
    lines:
      - ''
      - '&a&l[Mudah] &f%objective_displayname_daily_easy%'
      - '&7  Progress Anda: &a%objective_score_daily_easy%'
      - ''
      - '&e&l[Sedang] &f%objective_displayname_daily_med%'
      - '&7  Progress Anda: &e%objective_score_daily_med%'
      - ''
      - '&c&l[Sulit] &f%objective_displayname_daily_hard%'
      - '&7  Progress Anda: &c%objective_score_daily_hard%'
      - ''
      - '&6=========================='
`
	
	cmd := fmt.Sprintf("cat << 'EOF' > /home/mcserv/server-minecraft/plugins/SimpleScore/scoreboards.yml\n%s\nEOF", yamlContent)
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Fatalf("Failed to run SSH command: %s, output: %s", err, string(out))
	}
	
	fmt.Println("SSH update complete! Reloading SimpleScore via RCON...")
	runRcon("simplescore reload")
	fmt.Println("All done!")
}
