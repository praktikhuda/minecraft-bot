package main

import (
	"fmt"
	"log"
	"golang.org/x/crypto/ssh"
)

func main() {
	config := &ssh.ClientConfig{
		User: "minecraft",
		Auth: []ssh.AuthMethod{
			ssh.Password("minecraft12345"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "147.93.159.183:22", config)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	cmd := `mcrcon -H 127.0.0.1 -p RahasiaMinecraft123 "list"`
	out, _ := session.CombinedOutput(cmd)
	fmt.Println("List output:", string(out))

	// Assuming a player named 'dhionk' or someone is online. We can just test a random online player or just run list first.
}
