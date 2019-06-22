package cmd

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/twitchyliquid64/subnet/subnet"
	"github.com/twitchyliquid64/subnet/subnet/cert"
	"go.uber.org/zap"
)

const (
	caCertPath = "ca.crt"
	caKeyPath  = "ca.key"

	serverCertPath = "server.crt"
	serverKeyPath  = "server.key"

	clientCertPath = "client.crt"
	clientKeyPath  = "client.key"

	listenAddr = "0.0.0.0"
	listenPort = "3234"

	defaultNetwork = "192.168.69.1/24"
)

// Command is the server command for subnet VPN.
type Command struct {
	Logger  *zap.Logger
	Network string
	IName   string

	tempDir string
	server  *subnet.Server
}

// New returns a new Command object.
func New(logger *zap.Logger) *Command {
	return &Command{
		Logger:  logger,
		Network: defaultNetwork,
	}
}

// Execute performs the server program.
func (c *Command) Execute() error {
	defer c.close()
	if err := c.ensureTempDir(); err != nil {
		return nil
	}
	if err := c.ensureCerts(); err != nil {
		return err
	}
	if err := c.runServer(); err != nil {
		return err
	}
	c.waitSignal()
	return nil
}

func (c *Command) close() {
	if c.tempDir != "" {
		if err := os.RemoveAll(c.tempDir); err != nil {
			c.Logger.Error("failed to remove temp dir", zap.Error(err))
		}
		c.Logger.Debug("remove temp dir", zap.String("path", c.tempDir))
	}
	if c.server != nil {
		if err := c.server.Close(); err != nil {
			c.Logger.Error("failed to close server", zap.Error(err))
		}
		c.Logger.Info("closed vpn server")
	}
}

func (c *Command) ensureTempDir() (err error) {
	c.tempDir, err = ioutil.TempDir("", "kubevpn")
	if err != nil {
		return err
	}
	c.Logger.Debug("create temp dir", zap.String("path", c.tempDir))
	return
}

func (c *Command) ensureCerts() error {
	if err := cert.MakeServerCert(serverCertPath, serverKeyPath, caCertPath, caKeyPath); err != nil {
		return err
	}
	return cert.IssueClientCert(caCertPath, caKeyPath, clientCertPath, clientKeyPath)
}

func (c *Command) runServer() (err error) {
	c.server, err = subnet.NewServer(listenAddr, listenPort, c.Network, c.IName, serverCertPath, serverKeyPath, caCertPath)
	if err != nil {
		return err
	}
	c.server.Run()
	c.Logger.Info("started vpn server")
	return
}

func (c *Command) waitSignal() {
	sig := make(chan os.Signal, 2)
	done := make(chan struct{}, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sig
		c.Logger.Info("caught a signal", zap.String("signal", s.String()))
		done <- struct{}{}
	}()
	<-done
}
