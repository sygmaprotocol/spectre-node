// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/handlers"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/spectre-node/mock"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	consensus "github.com/umbracle/go-eth-consensus"
	"go.uber.org/mock/gomock"
)

func SliceTo32Bytes(in []byte) [32]byte {
	var res [32]byte
	copy(res[:], in)
	return res
}

type StepHandlerTestSuite struct {
	suite.Suite

	depositHandler *handlers.StepEventHandler

	msgChan          chan []*message.Message
	mockEventFetcher *mock.MockEventFetcher
	mockStepProver   *mock.MockProver
	mockBlockFetcher *mock.MockBlockFetcher

	sourceDomain uint8
}

func TestRunConfigTestSuite(t *testing.T) {
	suite.Run(t, new(StepHandlerTestSuite))
}

func (s *StepHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockEventFetcher = mock.NewMockEventFetcher(ctrl)
	s.mockStepProver = mock.NewMockProver(ctrl)
	s.mockBlockFetcher = mock.NewMockBlockFetcher(ctrl)
	s.msgChan = make(chan []*message.Message, 10)
	s.sourceDomain = 1
	s.depositHandler = handlers.NewStepEventHandler(
		s.msgChan,
		s.mockEventFetcher,
		s.mockBlockFetcher,
		s.mockStepProver,
		common.HexToAddress("0xb0b13f0109ef097C3Aa70Fb543EA4942114A845d"),
		s.sourceDomain,
		[]uint8{1, 2, 3})
}

func (s *StepHandlerTestSuite) Test_HandleEvents_FetchingArgsFails() {
	s.mockStepProver.EXPECT().StepArgs().Return(nil, fmt.Errorf("Error"))

	err := s.depositHandler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: phase0.Epoch(1024),
		},
	})
	s.NotNil(err)

	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *StepHandlerTestSuite) Test_HandleEvents_FetchingLogsFails() {
	s.mockStepProver.EXPECT().StepArgs().Return(&prover.StepArgs{
		Update: &consensus.LightClientFinalityUpdateDeneb{
			FinalizedHeader: &consensus.LightClientHeaderDeneb{
				Header: &consensus.BeaconBlockHeader{
					Slot: 10,
				},
			},
		},
	}, nil)
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(nil, fmt.Errorf("error"))

	err := s.depositHandler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: phase0.Epoch(1024),
		},
	})
	s.NotNil(err)

	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *StepHandlerTestSuite) Test_HandleEvents_FirstStep_StepExecuted() {
	s.mockStepProver.EXPECT().StepArgs().Return(&prover.StepArgs{
		Update: &consensus.LightClientFinalityUpdateDeneb{
			FinalizedHeader: &consensus.LightClientHeaderDeneb{
				Header: &consensus.BeaconBlockHeader{
					Slot: 10,
				},
				Execution: &consensus.ExecutionPayloadHeaderDeneb{},
			},
		},
	}, nil)
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Body: &deneb.BeaconBlockBody{
						ExecutionPayload: &deneb.ExecutionPayload{
							BlockNumber: 100,
						},
					},
				},
			},
		},
	}, nil)
	s.mockStepProver.EXPECT().StepProof(gomock.Any()).Return(&prover.EvmProof[evmMessage.SyncStepInput]{}, nil)

	err := s.depositHandler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: phase0.Epoch(1024),
		},
	})
	s.Nil(err)

	msgs, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(len(msgs), 1)
	s.Equal(msgs[0].Destination, uint8(2))
	msgs, err = readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(len(msgs), 1)
	s.Equal(msgs[0].Destination, uint8(3))
}

func (s *StepHandlerTestSuite) Test_HandleEvents_SecondStep_MissingDeposits() {
	s.mockStepProver.EXPECT().StepArgs().Return(&prover.StepArgs{
		Update: &consensus.LightClientFinalityUpdateDeneb{
			FinalizedHeader: &consensus.LightClientHeaderDeneb{
				Header: &consensus.BeaconBlockHeader{
					Slot: 10,
				},
				Execution: &consensus.ExecutionPayloadHeaderDeneb{},
			},
		},
	}, nil)
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Body: &deneb.BeaconBlockBody{
						ExecutionPayload: &deneb.ExecutionPayload{
							BlockNumber: 100,
						},
					},
				},
			},
		},
	}, nil)
	s.mockStepProver.EXPECT().StepProof(gomock.Any()).Return(&prover.EvmProof[evmMessage.SyncStepInput]{}, nil)
	err := s.depositHandler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: phase0.Epoch(1024),
		},
	})
	s.Nil(err)
	_, err = readFromChannel(s.msgChan)
	s.Nil(err)
	_, err = readFromChannel(s.msgChan)
	s.Nil(err)

	s.mockStepProver.EXPECT().StepArgs().Return(&prover.StepArgs{
		Update: &consensus.LightClientFinalityUpdateDeneb{
			FinalizedHeader: &consensus.LightClientHeaderDeneb{
				Header: &consensus.BeaconBlockHeader{
					Slot: 10,
				},
				Execution: &consensus.ExecutionPayloadHeaderDeneb{},
			},
		},
	}, nil)
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Body: &deneb.BeaconBlockBody{
						ExecutionPayload: &deneb.ExecutionPayload{
							BlockNumber: 110,
						},
					},
				},
			},
		},
	}, nil)
	s.mockEventFetcher.EXPECT().FetchEventLogs(context.Background(), gomock.Any(), gomock.Any(), big.NewInt(100), big.NewInt(110)).Return([]types.Log{}, nil)

	err = s.depositHandler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: phase0.Epoch(1024),
		},
	})
	s.Nil(err)

	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *StepHandlerTestSuite) Test_HandleEvents_SecondStep_ValidDeposits() {
	s.mockStepProver.EXPECT().StepArgs().Return(&prover.StepArgs{
		Update: &consensus.LightClientFinalityUpdateDeneb{
			FinalizedHeader: &consensus.LightClientHeaderDeneb{
				Header: &consensus.BeaconBlockHeader{
					Slot: 10,
				},
				Execution: &consensus.ExecutionPayloadHeaderDeneb{},
			},
		},
	}, nil)
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Body: &deneb.BeaconBlockBody{
						ExecutionPayload: &deneb.ExecutionPayload{
							BlockNumber: 100,
						},
					},
				},
			},
		},
	}, nil)
	s.mockStepProver.EXPECT().StepProof(gomock.Any()).Return(&prover.EvmProof[evmMessage.SyncStepInput]{}, nil)
	err := s.depositHandler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: phase0.Epoch(1024),
		},
	})
	s.Nil(err)
	_, err = readFromChannel(s.msgChan)
	s.Nil(err)
	_, err = readFromChannel(s.msgChan)
	s.Nil(err)

	s.mockStepProver.EXPECT().StepArgs().Return(&prover.StepArgs{
		Update: &consensus.LightClientFinalityUpdateDeneb{
			FinalizedHeader: &consensus.LightClientHeaderDeneb{
				Header: &consensus.BeaconBlockHeader{
					Slot: 10,
				},
				Execution: &consensus.ExecutionPayloadHeaderDeneb{},
			},
		},
	}, nil)
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Body: &deneb.BeaconBlockBody{
						ExecutionPayload: &deneb.ExecutionPayload{
							BlockNumber: 110,
						},
					},
				},
			},
		},
	}, nil)
	validDepositData, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000")
	s.mockEventFetcher.EXPECT().FetchEventLogs(context.Background(), gomock.Any(), gomock.Any(), big.NewInt(100), big.NewInt(110)).Return([]types.Log{
		{
			Data: validDepositData,
			Topics: []common.Hash{
				{},
				common.HexToHash("0xd68eb9b5E135b96c1Af165e1D8c4e2eB0E1CE4CD"),
			},
		},
	}, nil)
	s.mockStepProver.EXPECT().StepProof(gomock.Any()).Return(&prover.EvmProof[evmMessage.SyncStepInput]{}, nil)

	err = s.depositHandler.HandleEvents(&apiv1.Finality{
		Finalized: &phase0.Checkpoint{
			Epoch: phase0.Epoch(1024),
		},
	})
	s.Nil(err)

	msgs, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(len(msgs), 1)
	s.Equal(msgs[0].Destination, uint8(2))
	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}
