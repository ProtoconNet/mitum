package cmds

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spikeekips/mitum/launch/deploy"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/util"
	"golang.org/x/xerrors"
)

type DeployKeyNewCommand struct {
	*baseDeployKeyCommand
}

func NewDeployKeyNewCommand() DeployKeyNewCommand {
	return DeployKeyNewCommand{
		baseDeployKeyCommand: newBaseDeployKeyCommand("deploy-key-new"),
	}
}

func (cmd *DeployKeyNewCommand) Run(version util.Version) error {
	if err := cmd.Initialize(cmd, version); err != nil {
		return xerrors.Errorf("failed to initialize command: %w", err)
	}

	if err := cmd.requestToken(); err != nil {
		var pr network.Problem
		if xerrors.As(err, &pr) {
			cmd.Log().Error().Interface("problem", pr).Msg("failed")
		}

		return err
	}

	if err := cmd.requestNewKey(); err != nil {
		var pr network.Problem
		if xerrors.As(err, &pr) {
			cmd.Log().Error().Interface("problem", pr).Msg("failed")
		}

		return err
	}

	cmd.Log().Info().Msg("new deploy key")

	return nil
}

func (cmd *DeployKeyNewCommand) requestNewKey() error { // nolint:dupl
	res, c, err := cmd.requestWithToken(deploy.QuicHandlerPathDeployKeyNew, "GET")
	if err != nil {
		return xerrors.Errorf("failed to request new deploy key: %w", err)
	}
	defer func() {
		_ = c()
		_ = res.Body.Close()
	}()

	if res.StatusCode == http.StatusCreated {
		i, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return xerrors.Errorf("failed to read body: %w", err)
		}
		_, _ = fmt.Fprintln(os.Stdout, string(i))

		return nil
	}

	if i, err := network.LoadProblemFromResponse(res); err == nil {
		cmd.Log().Debug().Interface("response", res).Interface("problem", i).Msg("failed to request")

		return i
	}

	cmd.Log().Debug().Interface("response", res).Msg("failed to request")

	return xerrors.Errorf("failed to request deploy keys")
}