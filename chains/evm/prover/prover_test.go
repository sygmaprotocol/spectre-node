// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package prover_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/spectre-node/mock"
	consensus "github.com/umbracle/go-eth-consensus"
	"go.uber.org/mock/gomock"
)

type ProverTestSuite struct {
	suite.Suite

	prover           *prover.Prover
	lightClientMock  *mock.MockLightClient
	proverClientMock *mock.MockProverClient
	beaconClientMock *mock.MockBeaconClient
}

func TestRunProverTestSuite(t *testing.T) {
	suite.Run(t, new(ProverTestSuite))
}

func (s *ProverTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.lightClientMock = mock.NewMockLightClient(ctrl)
	s.proverClientMock = mock.NewMockProverClient(ctrl)
	s.beaconClientMock = mock.NewMockBeaconClient(ctrl)
	s.prover = prover.NewProver(s.proverClientMock, s.beaconClientMock, s.lightClientMock, prover.MAINNET_SPEC, 256)
}

func (s *ProverTestSuite) Test_StepProof_InvalidArgs() {
	s.lightClientMock.EXPECT().FinalityUpdate().Return(nil, fmt.Errorf("error"))

	_, err := s.prover.StepProof()

	s.NotNil(err)
}

func (s *ProverTestSuite) Test_StepProof_InvalidBlockRoot() {
	s.lightClientMock.EXPECT().FinalityUpdate().Return(&consensus.LightClientFinalityUpdateCapella{
		FinalizedHeader: &consensus.LightClientHeaderCapella{
			Header: &consensus.BeaconBlockHeader{
				Slot: 10,
			},
		},
	}, nil)
	s.beaconClientMock.EXPECT().BeaconBlockRoot(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))

	_, err := s.prover.StepProof()

	s.NotNil(err)
}

func (s *ProverTestSuite) Test_StepProof_InvalidBoostrap() {
	s.lightClientMock.EXPECT().FinalityUpdate().Return(&consensus.LightClientFinalityUpdateCapella{
		FinalizedHeader: &consensus.LightClientHeaderCapella{
			Header: &consensus.BeaconBlockHeader{
				Slot: 10,
			},
		},
	}, nil)
	s.beaconClientMock.EXPECT().BeaconBlockRoot(gomock.Any(), gomock.Any()).Return(&api.Response[*phase0.Root]{
		Data: &phase0.Root{},
	}, nil)
	s.lightClientMock.EXPECT().Bootstrap(gomock.Any()).Return(nil, fmt.Errorf("error"))

	_, err := s.prover.StepProof()

	s.NotNil(err)
}

func (s *ProverTestSuite) Test_StepProof_InvalidProof() {
	s.lightClientMock.EXPECT().FinalityUpdate().Return(&consensus.LightClientFinalityUpdateCapella{
		FinalizedHeader: &consensus.LightClientHeaderCapella{
			Header: &consensus.BeaconBlockHeader{
				Slot: 10,
			},
			Execution:       &consensus.ExecutionPayloadHeaderCapella{},
			ExecutionBranch: [4][32]byte{},
		},
		FinalityBranch: [][32]byte{{}, {}, {}, {}, {}, {}},
	}, nil)
	s.beaconClientMock.EXPECT().BeaconBlockRoot(gomock.Any(), gomock.Any()).Return(&api.Response[*phase0.Root]{
		Data: &phase0.Root{},
	}, nil)
	s.lightClientMock.EXPECT().Bootstrap(gomock.Any()).Return(&consensus.LightClientBootstrapCapella{
		CurrentSyncCommittee: &consensus.SyncCommittee{
			PubKeys: [512][48]byte{},
		},
	}, nil)
	s.beaconClientMock.EXPECT().Domain(gomock.Any(), gomock.Any(), gomock.Any()).Return(phase0.Domain{}, nil)
	s.proverClientMock.EXPECT().CallFor(context.Background(), gomock.Any(), "genEvmProofAndInstancesStepSyncCircuitWithWitness", gomock.Any()).Return(fmt.Errorf("error"))

	_, err := s.prover.StepProof()

	s.NotNil(err)
}

func (s *ProverTestSuite) Test_StepProof_ValidProof() {
	update := &consensus.LightClientFinalityUpdateCapella{
		FinalizedHeader: &consensus.LightClientHeaderCapella{
			Header: &consensus.BeaconBlockHeader{
				Slot: 10,
			},
			Execution: &consensus.ExecutionPayloadHeaderCapella{},
		},
		AttestedHeader: &consensus.LightClientHeaderCapella{
			Header: &consensus.BeaconBlockHeader{
				Slot: 11,
			},
		},
		SyncAggregate: &consensus.SyncAggregate{
			SyncCommiteeBits: [64]byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		FinalityBranch: [][32]byte{{}, {}, {}, {}, {}, {}},
	}
	s.lightClientMock.EXPECT().FinalityUpdate().Return(update, nil)
	s.beaconClientMock.EXPECT().BeaconBlockRoot(gomock.Any(), gomock.Any()).Return(&api.Response[*phase0.Root]{
		Data: &phase0.Root{},
	}, nil)
	s.lightClientMock.EXPECT().Bootstrap(gomock.Any()).Return(&consensus.LightClientBootstrapCapella{
		CurrentSyncCommittee: &consensus.SyncCommittee{
			PubKeys: [512][48]byte{{1}},
		},
	}, nil)
	s.beaconClientMock.EXPECT().Domain(gomock.Any(), gomock.Any(), gomock.Any()).Return(phase0.Domain{}, nil)
	s.proverClientMock.EXPECT().CallFor(context.Background(), gomock.Any(), "genEvmProofAndInstancesStepSyncCircuitWithWitness", gomock.Any()).DoAndReturn(func(ctx context.Context, reply *prover.ProverResponse, method string, args ...interface{}) error {
		*reply = prover.ProverResponse{
			Proof: []uint16{1, 2, 3},
		}
		return nil
	})

	_, err := s.prover.StepProof()

	s.Nil(err)
}

func (s *ProverTestSuite) Test_RotateProof_InvalidArgs() {
	s.lightClientMock.EXPECT().Updates(gomock.Any()).Return([]*consensus.LightClientUpdateCapella{}, fmt.Errorf("error"))

	_, err := s.prover.RotateProof(1000)

	s.NotNil(err)
}

func (s *ProverTestSuite) Test_RotateProof_InvalidSyncCommitteePoseidon() {
	s.lightClientMock.EXPECT().Updates(uint64(39)).Return([]*consensus.LightClientUpdateCapella{{
		FinalityBranch: [][32]byte{{1}},
		NextSyncCommittee: &consensus.SyncCommittee{
			PubKeys: [512][48]byte{},
		},
	}}, nil)

	s.proverClientMock.EXPECT().CallFor(context.Background(), gomock.Any(), "syncCommitteePoseidonCompressed", gomock.Any()).Return(fmt.Errorf("error"))

	_, err := s.prover.RotateProof(10000)

	s.NotNil(err)
}

func (s *ProverTestSuite) Test_RotateProof_InvalidProof() {
	s.lightClientMock.EXPECT().Updates(uint64(39)).Return([]*consensus.LightClientUpdateCapella{{
		FinalityBranch: [][32]byte{{1}, {}, {}, {}, {}, {}},
		NextSyncCommittee: &consensus.SyncCommittee{
			PubKeys: [512][48]byte{{}, {}, {}, {}, {}},
		},
		NextSyncCommitteeBranch: [][32]byte{{}, {}, {}, {}, {}},
	}}, nil)
	s.proverClientMock.EXPECT().CallFor(context.Background(), gomock.Any(), "syncCommitteePoseidonCompressed", gomock.Any()).DoAndReturn(func(ctx context.Context, resp *prover.CommitmentResponse, method string, args ...interface{}) error {
		*resp = prover.CommitmentResponse{
			Commitment: [32]byte{},
		}
		return nil
	})
	s.proverClientMock.EXPECT().CallFor(context.Background(), gomock.Any(), "genEvmProofAndInstancesRotationCircuitWithWitness", gomock.Any()).Return(fmt.Errorf("error"))

	_, err := s.prover.RotateProof(10000)

	s.NotNil(err)
}

func (s *ProverTestSuite) Test_RotateProof_ValidProof() {
	s.lightClientMock.EXPECT().Updates(uint64(39)).Return([]*consensus.LightClientUpdateCapella{{
		FinalityBranch: [][32]byte{{1}, {}, {}, {}, {}, {}},
		NextSyncCommittee: &consensus.SyncCommittee{
			PubKeys: [512][48]byte{{}, {}, {}, {}, {}},
		},
		NextSyncCommitteeBranch: [][32]byte{{}, {}, {}, {}, {}},
	}}, nil)
	s.proverClientMock.EXPECT().CallFor(context.Background(), gomock.Any(), "syncCommitteePoseidonCompressed", gomock.Any()).DoAndReturn(func(ctx context.Context, resp *prover.CommitmentResponse, method string, args ...interface{}) error {
		*resp = prover.CommitmentResponse{
			Commitment: [32]byte{},
		}
		return nil
	})
	s.proverClientMock.EXPECT().CallFor(context.Background(), gomock.Any(), "genEvmProofAndInstancesRotationCircuitWithWitness", gomock.Any()).DoAndReturn(func(ctx context.Context, resp *prover.ProverResponse, method string, args ...interface{}) error {
		*resp = prover.ProverResponse{
			Proof: []uint16{2, 3, 4},
		}
		return nil
	})

	_, err := s.prover.RotateProof(10000)

	s.Nil(err)
}
