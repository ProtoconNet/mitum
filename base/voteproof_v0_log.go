package base

import (
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"
)

func (vp VoteproofV0) MarshalLog(key string, e logging.Emitter, verbose bool) logging.Emitter {
	if !verbose {
		ev := logging.Dict().
			Hinted("height", vp.height).
			Hinted("round", vp.round).
			Hinted("stage", vp.stage).
			Bool("is_cloed", vp.closed).
			Str("result", vp.result.String()).
			Int("number_of_votes", len(vp.votes)).
			Int("number_of_ballots", len(vp.ballots))

		if vp.IsFinished() {
			ev = ev.Hinted("fact", vp.majority.Hash()).
				Time("finished_at", vp.finishedAt)
		}

		return e.Dict(key, ev)
	}

	r, _ := util.JSONMarshal(vp)

	return e.RawJSON(key, r)
}