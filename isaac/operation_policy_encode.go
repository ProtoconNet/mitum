package isaac

import (
	"time"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/key"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (po *PolicyOperationBodyV0) unpack(
	thresholdRatio base.ThresholdRatio,
	timeoutWaitingProposal time.Duration,
	intervalBroadcastingINITBallot time.Duration,
	intervalBroadcastingProposal time.Duration,
	waitBroadcastingACCEPTBallot time.Duration,
	intervalBroadcastingACCEPTBallot time.Duration,
	numberOfActingSuffrageNodes uint,
	timespanValidBallot,
	timeoutProcessProposal time.Duration,
) error {
	po.thresholdRatio = thresholdRatio
	po.timeoutWaitingProposal = timeoutWaitingProposal
	po.intervalBroadcastingINITBallot = intervalBroadcastingINITBallot
	po.intervalBroadcastingProposal = intervalBroadcastingProposal
	po.waitBroadcastingACCEPTBallot = waitBroadcastingACCEPTBallot
	po.intervalBroadcastingACCEPTBallot = intervalBroadcastingACCEPTBallot
	po.numberOfActingSuffrageNodes = numberOfActingSuffrageNodes
	po.timespanValidBallot = timespanValidBallot
	po.timeoutProcessProposal = timeoutProcessProposal

	return nil
}

func (spo *SetPolicyOperationV0) unpack(
	enc encoder.Encoder,
	h,
	factHash valuehash.Hash,
	factSignature key.Signature,
	bSigner,
	token,
	bPolicies []byte,
) error {
	var err error

	var signer key.Publickey
	if signer, err = key.DecodePublickey(enc, bSigner); err != nil {
		return err
	}

	var body PolicyOperationBodyV0
	if err := enc.Decode(bPolicies, &body); err != nil {
		return err
	}

	spo.h = h
	spo.factHash = factHash
	spo.factSignature = factSignature
	spo.SetPolicyOperationFactV0 = SetPolicyOperationFactV0{
		PolicyOperationBodyV0: body,
		signer:                signer,
		token:                 token,
	}

	return nil
}
