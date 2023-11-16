// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/events/handlers"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/spectre-node/mock"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"go.uber.org/mock/gomock"
)

type RotateHandlerTestSuite struct {
	suite.Suite

	handler *handlers.RotateHandler

	msgChan                  chan []*message.Message
	mockProver               *mock.MockProver
	mockSyncCommitteeFetcher *mock.MockSyncCommitteeFetcher
}

func TestRunRotateTestSuite(t *testing.T) {
	suite.Run(t, new(RotateHandlerTestSuite))
}

func (s *RotateHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockProver = mock.NewMockProver(ctrl)
	s.mockSyncCommitteeFetcher = mock.NewMockSyncCommitteeFetcher(ctrl)
	s.msgChan = make(chan []*message.Message, 2)
	s.handler = handlers.NewRotateHandler(
		s.msgChan,
		s.mockSyncCommitteeFetcher,
		s.mockProver,
		1,
		[]uint8{1, 2, 3},
	)
}

func (s *RotateHandlerTestSuite) Test_HandleEvents_FetchingCommitteeFails() {
	s.mockSyncCommitteeFetcher.EXPECT().SyncCommittee(context.Background(), gomock.Any()).Return(nil, fmt.Errorf("error"))

	err := s.handler.HandleEvents(&apiv1.Finality{
		Justified: &phase0.Checkpoint{
			Root: phase0.Root{},
		},
	})
	s.NotNil(err)

	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *RotateHandlerTestSuite) Test_HandleEvents_SyncCommitteeNotChanged() {
	s.mockSyncCommitteeFetcher.EXPECT().SyncCommittee(context.Background(), gomock.Any()).Return(&api.Response[*apiv1.SyncCommittee]{
		Data: &apiv1.SyncCommittee{},
	}, nil)

	err := s.handler.HandleEvents(&apiv1.Finality{
		Justified: &phase0.Checkpoint{
			Root: phase0.Root{},
		},
	})
	s.Nil(err)

	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *RotateHandlerTestSuite) Test_HandleEvents_NewSyncCommittee_ProofFails() {
	s.mockSyncCommitteeFetcher.EXPECT().SyncCommittee(context.Background(), gomock.Any()).Return(&api.Response[*apiv1.SyncCommittee]{
		Data: &apiv1.SyncCommittee{
			Validators: []phase0.ValidatorIndex{128},
		},
	}, nil)
	s.mockProver.EXPECT().StepProof().Return(nil, fmt.Errorf("error"))

	err := s.handler.HandleEvents(&apiv1.Finality{
		Justified: &phase0.Checkpoint{
			Root: phase0.Root{},
		},
	})
	s.NotNil(err)
	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)

	s.mockSyncCommitteeFetcher.EXPECT().SyncCommittee(context.Background(), gomock.Any()).Return(&api.Response[*apiv1.SyncCommittee]{
		Data: &apiv1.SyncCommittee{
			Validators: []phase0.ValidatorIndex{128},
		},
	}, nil)
	s.mockProver.EXPECT().StepProof().Return(&prover.EvmProof{}, nil)
	s.mockProver.EXPECT().RotateProof(uint64(100)).Return(nil, fmt.Errorf("error"))

	err = s.handler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: 100,
		},
	})
	s.NotNil(err)
	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *RotateHandlerTestSuite) Test_HandleEvents_NewSyncCommittee() {
	s.mockSyncCommitteeFetcher.EXPECT().SyncCommittee(context.Background(), gomock.Any()).Return(&api.Response[*apiv1.SyncCommittee]{
		Data: &apiv1.SyncCommittee{
			Validators: []phase0.ValidatorIndex{128},
		},
	}, nil)
	s.mockProver.EXPECT().StepProof().Return(&prover.EvmProof{}, nil)
	s.mockProver.EXPECT().RotateProof(uint64(100)).Return(&prover.EvmProof{}, nil)

	err := s.handler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: 100,
		},
	})
	s.Nil(err)

	msgs, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(msgs[0].Destination, uint8(2))

	msgs, err = readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(msgs[0].Destination, uint8(3))
}
