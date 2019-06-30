package stage

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/docker/docker/pkg/term"
	vbox "github.com/oclaussen/dodo/pkg/stage/virtualbox"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

func (stage *Stage) RunSSHCommand(command string) (string, error) {
	opts, err := vbox.GetSSHOptions(stage.name)
	if err != nil {
		return "", err
	}

	key, err := ioutil.ReadFile(filepath.Join(stage.stateDir, "machines", stage.name, "id_rsa"))
	if err != nil {
		return "", err
	}

	privateKey, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", err
	}

	authMethods := []ssh.AuthMethod{ssh.PublicKeys(privateKey)}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(opts.Hostname, strconv.Itoa(opts.Port)), &ssh.ClientConfig{
		User:            opts.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return "", errors.Wrap(err, "could not connect to SSH")
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", nil
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	return string(output), err
}

func (stage *Stage) SSH() error {
	currentStatus, err := vbox.GetStatus(stage.name)
	if err != nil {
		return err
	}

	if currentStatus != vbox.Running {
		return errors.New("stage is not up")
	}

	opts, err := vbox.GetSSHOptions(stage.name)
	if err != nil {
		return err
	}

	key, err := ioutil.ReadFile(filepath.Join(stage.stateDir, "machines", stage.name, "id_rsa"))
	if err != nil {
		return err
	}

	privateKey, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return err
	}

	authMethods := []ssh.AuthMethod{ssh.PublicKeys(privateKey)}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(opts.Hostname, strconv.Itoa(opts.Port)), &ssh.ClientConfig{
		User:            opts.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return errors.Wrap(err, "could not connect to SSH")
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var (
		termWidth, termHeight int
	)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
	}

	fd := os.Stdin.Fd()

	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}

		defer term.RestoreTerminal(fd, oldState)

		winsize, err := term.GetWinsize(fd)
		if err != nil {
			termWidth = 80
			termHeight = 24
		} else {
			termWidth = int(winsize.Width)
			termHeight = int(winsize.Height)
		}
	}

	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		return err
	}
	if err := session.Shell(); err != nil {
		return err
	}
	if err := session.Wait(); err != nil {
		return err
	}

	return nil
}
