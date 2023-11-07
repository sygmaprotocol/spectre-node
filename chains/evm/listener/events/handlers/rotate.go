// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"math/big"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
)

type SyncCommitteeFetcher interface {
	SyncCommittee(ctx context.Context, opts *api.SyncCommitteeOpts) (*api.Response[*apiv1.SyncCommittee], error)
}

type RotateHandler struct {
	syncCommitteeFetcher SyncCommitteeFetcher

	currentSyncCommittee *api.Response[*apiv1.SyncCommittee]
}

func NewRotateHandler(syncCommitteeFetcher SyncCommitteeFetcher) *RotateHandler {
	return &RotateHandler{
		syncCommitteeFetcher: syncCommitteeFetcher,
	}
}

// HandleEvents checks if there is a new sync committee and sends a rotate message
// if there is
func (h *RotateHandler) HandleEvents(startBlock *big.Int, endBlock *big.Int) error {
	syncCommittee, err := h.syncCommitteeFetcher.SyncCommittee(context.Background(), &api.SyncCommitteeOpts{
		State: "finalized",
	})
	if err != nil {
		return err
	}

	if syncCommittee.Data.String() == h.currentSyncCommittee.Data.String() {
		return nil
	}

	return nil
}
