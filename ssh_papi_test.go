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
	out, _ := session.CombinedOutput(`mcrcon -H 127.0.0.1 -P 25575 -p "RahasiaMinecraft123" "placeholderapi parse MrPheee %objective_displayname_{daily_easy}%"`)
	fmt.Println("With Braces:", string(out))
	session.Close()
	
	session2, _ := client.NewSession()
	out2, _ := session2.CombinedOutput(`mcrcon -H 127.0.0.1 -P 25575 -p "RahasiaMinecraft123" "placeholderapi parse MrPheee %objective_displayname_daily_easy%"`)
	fmt.Println("No Braces:", string(out2))
	session2.Close()
}
