package client

import (
	"context"
	"errors"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewSSHClient(addr, user, password string) (client *ssh.Client, err error) {
	config := &ssh.ClientConfig{
		User: user,
		// https://github.com/golang/go/issues/19767
		// as clientConfig is non-permissive by default
		// you can set ssh.InsercureIgnoreHostKey to allow any host
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}
	// Connect
	client, err = ssh.Dial("tcp", net.JoinHostPort(addr, "22"), config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func SshRemoteRunCommandWithTimeout(sshClient *ssh.Client, command string, timeout time.Duration) (string, error) {
	if timeout < 1 {
		return "", errors.New("timeout must be valid")
	}

	session, err := sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	resChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		// run shell script
		if output, err := session.CombinedOutput(command); err != nil {
			errChan <- err
		} else {
			resChan <- string(output)
		}
	}()

	select {
	case err := <-errChan:
		return "", err
	case ms := <-resChan:
		return ms, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
