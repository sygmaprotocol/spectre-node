// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package listener_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener"
	"github.com/sygmaprotocol/spectre-node/mock"
	"go.uber.org/mock/gomock"
)

type ListenerTestSuite struct {
	suite.Suite

	listener           *listener.EVMListener
	mockBeaconProvider *mock.MockBeaconProvider
	mockEventHandler   *mock.MockEventHandler
}

func TestRunListenerTestSuite(t *testing.T) {
	suite.Run(t, new(ListenerTestSuite))
}

func (s *ListenerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockBeaconProvider = mock.NewMockBeaconProvider(ctrl)
	s.mockEventHandler = mock.NewMockEventHandler(ctrl)

	s.listener = listener.NewEVMListener(
		s.mockBeaconProvider,
		[]listener.EventHandler{s.mockEventHandler, s.mockEventHandler},
		1,
		time.Millisecond*50,
		big.NewInt(32))
}

func (s *ListenerTestSuite) Test_ListenToEvents_CheckpointUnavailable() {
	s.mockBeaconProvider.EXPECT().Finality(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))

	ctx, cancel := context.WithCancel(context.Background())
	go s.listener.ListenToEvents(ctx, big.NewInt(0))

	time.Sleep(time.Millisecond * 25)
	cancel()
}

func (s *ListenerTestSuite) Test_ListenToEvents_CheckpointNotUpdated() {
	s.mockBeaconProvider.EXPECT().Finality(gomock.Any(), gomock.Any()).Return(&api.Response[*apiv1.Finality]{
		Data: &apiv1.Finality{
			Finalized: &phase0.Checkpoint{
				Root: phase0.Root([32]byte{}),
			},
		},
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	go s.listener.ListenToEvents(ctx, big.NewInt(0))

	time.Sleep(time.Millisecond * 25)
	cancel()
}

func (s *ListenerTestSuite) Test_ListenToEvents_FetchingBlockFails() {
	s.mockBeaconProvider.EXPECT().Finality(gomock.Any(), gomock.Any()).Return(&api.Response[*apiv1.Finality]{
		Data: &apiv1.Finality{
			Finalized: &phase0.Checkpoint{
				Root: phase0.Root([32]byte{1}),
			},
			Justified: &phase0.Checkpoint{
				Root: phase0.Root([32]byte{1}),
			},
		},
	}, nil)
	s.mockBeaconProvider.EXPECT().SignedBeaconBlock(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))

	ctx, cancel := context.WithCancel(context.Background())
	go s.listener.ListenToEvents(ctx, big.NewInt(0))

	time.Sleep(time.Millisecond * 25)
	cancel()
}

func (s *ListenerTestSuite) Test_ListenToEvents_ValidCheckpoint() {
	// First pass
	s.mockBeaconProvider.EXPECT().Finality(gomock.Any(), gomock.Any()).Return(&api.Response[*apiv1.Finality]{
		Data: &apiv1.Finality{
			Finalized: &phase0.Checkpoint{
				Root: phase0.Root([32]byte{1}),
			},
			Justified: &phase0.Checkpoint{
				Root: phase0.Root([32]byte{1}),
			},
		},
	}, nil)
	s.mockBeaconProvider.EXPECT().SignedBeaconBlock(gomock.Any(), gomock.Any()).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Capella: &capella.SignedBeaconBlock{
				Message: &capella.BeaconBlock{
					Body: &capella.BeaconBlockBody{
						ExecutionPayload: &capella.ExecutionPayload{
							BlockNumber: 100,
						},
					},
				},
			},
		},
	}, nil)
	s.mockEventHandler.EXPECT().HandleEvents(big.NewInt(68), big.NewInt(100)).Return(fmt.Errorf("error"))

	// Second pass
	s.mockBeaconProvider.EXPECT().Finality(gomock.Any(), gomock.Any()).Return(&api.Response[*apiv1.Finality]{
		Data: &apiv1.Finality{
			Finalized: &phase0.Checkpoint{
				Root: phase0.Root([32]byte{1}),
			},
			Justified: &phase0.Checkpoint{
				Root: phase0.Root([32]byte{1}),
			},
		},
	}, nil)
	s.mockBeaconProvider.EXPECT().SignedBeaconBlock(gomock.Any(), gomock.Any()).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Capella: &capella.SignedBeaconBlock{
				Message: &capella.BeaconBlock{
					Body: &capella.BeaconBlockBody{
						ExecutionPayload: &capella.ExecutionPayload{
							BlockNumber: 100,
						},
					},
				},
			},
		},
	}, nil)
	s.mockEventHandler.EXPECT().HandleEvents(big.NewInt(68), big.NewInt(100)).Return(nil)
	s.mockEventHandler.EXPECT().HandleEvents(big.NewInt(68), big.NewInt(100)).Return(nil)
	// Third pass
	s.mockBeaconProvider.EXPECT().Finality(gomock.Any(), gomock.Any()).Return(&api.Response[*apiv1.Finality]{
		Data: &apiv1.Finality{
			Finalized: &phase0.Checkpoint{
				Root: phase0.Root([32]byte{1}),
			},
			Justified: &phase0.Checkpoint{
				Root: phase0.Root([32]byte{1}),
			},
		},
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	go s.listener.ListenToEvents(ctx, big.NewInt(0))

	time.Sleep(time.Millisecond * 25)
	cancel()
}
