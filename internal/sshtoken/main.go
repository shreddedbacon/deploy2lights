package sshtoken

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

func publicKey(path string, skipAgent bool) (ssh.AuthMethod, func() error) {
	noopCloseFunc := func() error { return nil }

	if skipAgent != true {
		// Connect to SSH agent to ask for unencrypted private keys
		if sshAgentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
			sshAgent := agent.NewClient(sshAgentConn)

			keys, _ := sshAgent.List()
			if len(keys) > 0 {
				// There are key(s) in the agent
				//defer sshAgentConn.Close()
				return ssh.PublicKeysCallback(sshAgent.Signers), sshAgentConn.Close
			}
		}
	}

	key, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error was:", err.Error())
		os.Exit(1)
	}

	// Try to look for an unencrypted private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Println("Error was:", err.Error())
		fmt.Println("Lagoon CLI cannot decode private keys, you will need to add your private key to your ssh-agent.")
		os.Exit(1)
	} else if err == nil {
		// return unencrypted private key
		return ssh.PublicKeys(signer), noopCloseFunc
	}

	// Handle encrypted private keys
	fmt.Println("Found an encrypted private key!")
	fmt.Printf("Enter passphrase for '%s': ", path)
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	signer, err = ssh.ParsePrivateKeyWithPassphrase(key, bytePassword)
	return ssh.PublicKeys(signer), noopCloseFunc
}

func GetToken(userPath, sshHost, sshPort string) (string, error) {
	skipAgent := false
	privateKey := fmt.Sprintf("%s/.ssh/id_rsa", userPath)
	authMethod, closeSSHAgent := publicKey(privateKey, skipAgent)
	config := &ssh.ClientConfig{
		User: "lagoon",
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	defer closeSSHAgent()

	sshHostString := fmt.Sprintf("%s:%s",
		sshHost,
		sshPort,
	)
	conn, err := ssh.Dial("tcp", sshHostString, config)
	if err != nil {
		return "", fmt.Errorf("couldn't connect to %s: %v", sshHostString, err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("couldn't open session: %v", err)
	}

	out, err := session.CombinedOutput("token")
	if err != nil {
		return "", fmt.Errorf("couldn't get token: %v", err)
	}
	return strings.TrimSpace(string(out)), err
}
