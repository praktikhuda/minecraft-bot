package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
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

	session, _ := client.NewSession()
	out, _ := session.CombinedOutput(`echo 'mc12345' | sudo -S systemctl disable --now mcbot.service`)
	fmt.Println("DISABLE:", string(out))
	session.Close()
	
	// And restart the new bot one final time to cleanly reset quests
	session2, _ := client.NewSession()
	out2, _ := session2.CombinedOutput(`echo 'mc12345' | sudo -S systemctl restart mcgo.service`)
	fmt.Println("RESTART NEW:", string(out2))
	session2.Close()
}
