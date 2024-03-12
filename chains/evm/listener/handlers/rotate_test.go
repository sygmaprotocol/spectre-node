// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers_test

import (
	"fmt"
	"math/big"
	"testing"

	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/handlers"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/spectre-node/mock"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	consensus "github.com/umbracle/go-eth-consensus"
	"go.uber.org/mock/gomock"
)

func readFromChannel(msgChan chan []*message.Message) ([]*message.Message, error) {
	select {
	case msgs := <-msgChan:
		return msgs, nil
	default:
		return make([]*message.Message, 0), fmt.Errorf("no message sent")
	}
}

type RotateHandlerTestSuite struct {
	suite.Suite

	handler *handlers.RotateHandler

	msgChan          chan []*message.Message
	mockProver       *mock.MockProver
	mockPeriodStorer *mock.MockPeriodStorer
}

func TestRunRotateTestSuite(t *testing.T) {
	suite.Run(t, new(RotateHandlerTestSuite))
}

func (s *RotateHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockProver = mock.NewMockProver(ctrl)
	s.mockPeriodStorer = mock.NewMockPeriodStorer(ctrl)
	s.msgChan = make(chan []*message.Message, 2)
	s.handler = handlers.NewRotateHandler(
		s.msgChan,
		s.mockPeriodStorer,
		s.mockProver,
		1,
		[]uint8{2, 3},
		256,
		big.NewInt(3),
	)
}

func (s *RotateHandlerTestSuite) Test_HandleEvents_CurrentPeriodOlderThanLatest() {
	err := s.handler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: phase0.Epoch(765),
		},
	})
	s.Nil(err)

	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *RotateHandlerTestSuite) Test_HandleEvents_ValidPeriod() {
	s.mockPeriodStorer.EXPECT().StorePeriod(uint8(1), big.NewInt(4)).Return(nil)
	s.mockProver.EXPECT().RotateArgs(uint64(4)).Return(&prover.RotateArgs{
		Update:  &consensus.LightClientUpdateDeneb{},
		Domain:  phase0.Domain{},
		Spec:    "mainnet",
		Pubkeys: [512][48]byte{},
	}, nil)
	s.mockProver.EXPECT().RotateProof(gomock.Any()).Return(&prover.EvmProof[struct{}]{
		Proof: []byte{},
		Input: struct{}{},
	}, nil)
	s.mockProver.EXPECT().StepProof(gomock.Any()).Return(&prover.EvmProof[evmMessage.SyncStepInput]{
		Proof: []byte{},
		Input: evmMessage.SyncStepInput{},
	}, nil)

	err := s.handler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: phase0.Epoch(1024),
		},
	})
	s.Nil(err)

	msg1, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(len(msg1), 1)
	msg2, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(len(msg2), 1)
}
