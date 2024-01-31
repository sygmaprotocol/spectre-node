// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"math/big"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/rs/zerolog/log"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	consensus "github.com/umbracle/go-eth-consensus"
)

type SyncCommitteeFetcher interface {
	SyncCommittee(ctx context.Context, opts *api.SyncCommitteeOpts) (*api.Response[*apiv1.SyncCommittee], error)
}

type PeriodStorer interface {
	Period(domainID uint8) (*big.Int, error)
	StorePeriod(domainID uint8, period *big.Int) error
}

type RotateHandler struct {
	domainID uint8
	domains  []uint8
	msgChan  chan []*message.Message

	prover       Prover
	periodStorer PeriodStorer
	latestPeriod *big.Int

	committeePeriodLength uint64
}

func NewRotateHandler(
	msgChan chan []*message.Message,
	periodStorer PeriodStorer,
	prover Prover,
	domainID uint8,
	domains []uint8,
	committeePeriodLenght uint64,
	latestPeriod *big.Int,
) *RotateHandler {
	return &RotateHandler{
		prover:                prover,
		periodStorer:          periodStorer,
		domainID:              domainID,
		domains:               domains,
		msgChan:               msgChan,
		committeePeriodLength: committeePeriodLenght,
		latestPeriod:          latestPeriod,
	}
}

// HandleEvents checks if the current period is newer than the last stored
// period and rotates the committee if it is
func (h *RotateHandler) HandleEvents(checkpoint *apiv1.Finality) error {
	currentPeriod := uint64(checkpoint.Finalized.Epoch) / h.committeePeriodLength
	if currentPeriod <= h.latestPeriod.Uint64() {
		return nil
	}

	targetPeriod := new(big.Int).Add(h.latestPeriod, big.NewInt(1))
	args, err := h.prover.RotateArgs(targetPeriod.Uint64())
	if err != nil {
		return err
	}
	sArgs := &prover.StepArgs{
		Pubkeys: args.Pubkeys,
		Update: &consensus.LightClientFinalityUpdateCapella{
			AttestedHeader:  args.Update.AttestedHeader,
			FinalizedHeader: args.Update.FinalizedHeader,
			FinalityBranch:  args.Update.FinalityBranch,
			SyncAggregate:   args.Update.SyncAggregate,
			SignatureSlot:   args.Update.SignatureSlot,
		},
		Domain: args.Domain,
		Spec:   args.Spec,
	}

	log.Info().Uint8("domainID", h.domainID).Uint64("period", targetPeriod.Uint64()+1).Msgf("Rotating committee")

	rotateProof, err := h.prover.RotateProof(args)
	if err != nil {
		return err
	}
	stepProof, err := h.prover.StepProof(sArgs)
	if err != nil {
		return err
	}

	for _, domain := range h.domains {
		if domain == h.domainID {
			continue
		}

		log.Debug().Uint8("domainID", h.domainID).Msgf("Sending rotate message to domain %d", domain)
		h.msgChan <- []*message.Message{
			evmMessage.NewEvmRotateMessage(h.domainID, domain, evmMessage.RotateData{
				RotateInput: rotateProof.Input,
				RotateProof: rotateProof.Proof,
				StepProof:   stepProof.Proof,
				StepInput:   stepProof.Input,
			}),
		}
	}

	h.latestPeriod = targetPeriod
	return h.periodStorer.StorePeriod(h.domainID, targetPeriod)
}
