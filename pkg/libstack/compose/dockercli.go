package compose

import "github.com/docker/cli/cli/command"

type buildKitDisabledCli struct {
	command.Cli
}

func (c buildKitDisabledCli) BuildKitEnabled() (bool, error) {
	return false, nil
}
