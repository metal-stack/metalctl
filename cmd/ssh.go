package cmd

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/avast/retry-go/v4"
	"golang.org/x/net/proxy"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// SSHClient opens an interactive ssh session to the host on port with user, authenticated by the key.
func SSHClient(user, keyfile, host string, port int) error {
	sshConfig, err := getSSHConfig(user, keyfile)
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %w", err)
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	return createSSHSession(client)
}

func SSHClientOverSOCKS5(user, keyfile, host string, port int, proxyAddr string) error {
	sshConfig, err := getSSHConfig(user, keyfile)
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %w", err)
	}

	client, err := getProxiedSSHClient(fmt.Sprintf("%s:%d", host, port), proxyAddr, sshConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	return createSSHSession(client)
}

func getProxiedSSHClient(sshServerAddress, proxyAddr string, sshConfig *ssh.ClientConfig) (*ssh.Client, error) {
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create a proxy dialer: %w", err)
	}

	var conn net.Conn
	err = retry.Do(
		func() error {
			fmt.Printf(".")
			conn, err = dialer.Dial("tcp", sshServerAddress)
			if err == nil {
				fmt.Printf("\n")
				return nil
			}
			return err
		},
		retry.Attempts(50),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy at address %s: %w", proxyAddr, err)
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, sshServerAddress, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh connection: %w", err)
	}

	return ssh.NewClient(c, chans, reqs), nil
}

func createSSHSession(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Set IO
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	// Set up terminal modes
	// https://net-ssh.github.io/net-ssh/classes/Net/SSH/Connection/Term.html
	// https://www.ietf.org/rfc/rfc4254.txt
	// https://godoc.org/golang.org/x/crypto/ssh
	// THIS IS THE TITLE
	// https://pythonhosted.org/ANSIColors-balises/ANSIColors.html
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,      // enable echoing
		ssh.TTY_OP_ISPEED: 115200, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 115200, // output speed = 14.4kbaud
	}

	fileDescriptor := int(os.Stdin.Fd())

	if term.IsTerminal(fileDescriptor) {
		originalState, err := term.MakeRaw(fileDescriptor)
		if err != nil {
			return err
		}
		defer func() {
			err = term.Restore(fileDescriptor, originalState)
			if err != nil {
				fmt.Printf("error restoring ssh terminal:%v\n", err)
			}
		}()

		termWidth, termHeight, err := term.GetSize(fileDescriptor)
		if err != nil {
			return err
		}

		err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if err != nil {
			return err
		}
	}

	err = session.Shell()
	if err != nil {
		return err
	}

	// You should now be connected via SSH with a fully-interactive terminal
	// This call blocks until the user exits the session (e.g. via CTRL + D)
	return session.Wait()
}

func getSSHConfig(user, keyfile string) (*ssh.ClientConfig, error) {
	keyfile, err := expandFilepath(keyfile)
	if err != nil {
		return nil, err
	}

	publicKeyAuthMethod, err := publicKey(keyfile)
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			publicKeyAuthMethod,
		},
		//nolint:gosec
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}, nil
}

func publicKey(path string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
