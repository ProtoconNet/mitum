package channetwork

import (
	"context"

	"github.com/spikeekips/mitum/base/seal"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"
)

type Server struct {
	*logging.Logging
	*util.ContextDaemon
	newSealHandler network.NewSealHandler
	ch             *Channel
}

func NewServer(ch *Channel) *Server {
	sv := &Server{
		Logging: logging.NewLogging(func(c logging.Context) logging.Emitter {
			return c.Str("module", "network-chan-server")
		}),
		ch: ch,
	}

	sv.ContextDaemon = util.NewContextDaemon("network-chan-server", sv.run)

	return sv
}

func (*Server) Initialize() error {
	return nil
}

func (sv *Server) SetLogger(l logging.Logger) logging.Logger {
	_ = sv.Logging.SetLogger(l)
	_ = sv.ContextDaemon.SetLogger(l)

	return sv.Log()
}

func (*Server) SetHasSealHandler(network.HasSealHandler)   {}
func (*Server) SetGetSealsHandler(network.GetSealsHandler) {}

func (sv *Server) SetNewSealHandler(f network.NewSealHandler) {
	sv.newSealHandler = f
}

func (*Server) SetNodeInfoHandler(network.NodeInfoHandler)           {}
func (*Server) NodeInfoHandler() network.NodeInfoHandler             { return nil }
func (*Server) SetBlockDataMapsHandler(network.BlockDataMapsHandler) {}
func (*Server) SetBlockDataHandler(network.BlockDataHandler)         {}

func (sv *Server) run(ctx context.Context) error {
end:
	for {
		select {
		case <-ctx.Done():
			break end
		case sl := <-sv.ch.ReceiveSeal():
			go func(sl seal.Seal) {
				if sv.newSealHandler == nil {
					sv.Log().Error().Msg("no NewSealHandler")
					return
				}

				if err := sv.newSealHandler(sl); err != nil {
					seal.LogEventWithSeal(
						sl,
						sv.Log().Error().Err(err),
						sv.Log().IsVerbose(),
					).Msg("failed to receive new seal")

					return
				}
			}(sl)
		}
	}

	return nil
}
