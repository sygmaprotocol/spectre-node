// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/rs/zerolog/log"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	consensus "github.com/umbracle/go-eth-consensus"
)

type SyncCommitteeFetcher interface {
	SyncCommittee(ctx context.Context, opts *api.SyncCommitteeOpts) (*api.Response[*apiv1.SyncCommittee], error)
}

type RotateHandler struct {
	domainID uint8
	domains  []uint8
	msgChan  chan []*message.Message

	prover Prover

	syncCommitteeFetcher SyncCommitteeFetcher
	currentSyncCommittee *apiv1.SyncCommittee
}

func NewRotateHandler(msgChan chan []*message.Message, syncCommitteeFetcher SyncCommitteeFetcher, prover Prover, domainID uint8, domains []uint8) *RotateHandler {
	return &RotateHandler{
		syncCommitteeFetcher: syncCommitteeFetcher,
		prover:               prover,
		domainID:             domainID,
		domains:              domains,
		msgChan:              msgChan,
		currentSyncCommittee: &apiv1.SyncCommittee{},
	}
}

// HandleEvents checks if there is a new sync committee and sends a rotate message
// if there is
func (h *RotateHandler) HandleEvents(checkpoint *apiv1.Finality) error {
	syncCommittee, err := h.syncCommitteeFetcher.SyncCommittee(context.Background(), &api.SyncCommitteeOpts{
		State: "justified",
	})
	if err != nil {
		return err
	}
	if syncCommittee.Data.String() == h.currentSyncCommittee.String() {
		return nil
	}

	log.Info().Uint8("domainID", h.domainID).Msgf("Rotating committee")

	rArgs, err := h.prover.RotateArgs(uint64(checkpoint.Justified.Epoch))
	if err != nil {
		return err
	}
	sArgs, err := h.prover.StepArgs()
	if err != nil {
		return err
	}
	sArgs.Update = &consensus.LightClientFinalityUpdateCapella{
		AttestedHeader:  rArgs.Update.AttestedHeader,
		FinalizedHeader: rArgs.Update.FinalizedHeader,
		FinalityBranch:  rArgs.Update.FinalityBranch,
		SyncAggregate:   rArgs.Update.SyncAggregate,
		SignatureSlot:   rArgs.Update.SignatureSlot,
	}

	rotateProof, err := h.prover.RotateProof(rArgs)
	if err != nil {
		return err
	}
	stepProof, err := h.prover.StepProof(sArgs)
	if err != nil {
		return err
	}

	for _, domain := range h.domains {
		/*
			TODO: removed for testing purposes
			if domain == h.domainID {
				continue
			}
		*/

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

	return nil
}
