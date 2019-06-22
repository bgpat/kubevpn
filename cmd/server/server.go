package main

import (
	"github.com/bgpat/kubevpn/pkg/server/cmd"
	"go.uber.org/zap"
)

var logger, _ = zap.NewProduction()

func main() {
	command := cmd.New(logger)
	if err := command.Execute(); err != nil {
		logger.Fatal(err.Error())
	}
}
