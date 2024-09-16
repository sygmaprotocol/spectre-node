// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/events"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/events/handlers"
	"github.com/sygmaprotocol/spectre-node/mock"
	evmMessage "github.com/sygmaprotocol/sygma-core/relayer/message"
	"go.uber.org/mock/gomock"
)

type HashiHandlerTestSuite struct {
	suite.Suite

	hashiHandler *handlers.HashiDomainCollector

	msgChan          chan []*evmMessage.Message
	mockEventFetcher *mock.MockEventFetcher
	domains          []uint8
	sourceDomain     uint8
	yahoAddress      common.Address
}

func TestRunHashiHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HashiHandlerTestSuite))
}

func (s *HashiHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockEventFetcher = mock.NewMockEventFetcher(ctrl)
	s.msgChan = make(chan []*evmMessage.Message, 2)
	s.sourceDomain = 1
	s.yahoAddress = common.HexToAddress("0xa83114A443dA1CecEFC50368531cACE9F37fCCcb")
	s.hashiHandler = handlers.NewHashiDomainCollector(
		s.sourceDomain,
		s.yahoAddress,
		s.mockEventFetcher,
		s.domains,
	)
}

func (s *HashiHandlerTestSuite) Test_CollectDomains_FetchingLogFails() {
	s.mockEventFetcher.EXPECT().FetchEventLogs(gomock.Any(), s.yahoAddress, string(events.MessageDispatchedSig), big.NewInt(100), big.NewInt(1100)).Return([]types.Log{}, nil)
	s.mockEventFetcher.EXPECT().FetchEventLogs(gomock.Any(), s.yahoAddress, string(events.MessageDispatchedSig), big.NewInt(1101), big.NewInt(2101)).Return([]types.Log{{}}, fmt.Errorf("error"))

	_, err := s.hashiHandler.CollectDomains(big.NewInt(100), big.NewInt(2568))

	s.NotNil(err)
}

func (s *HashiHandlerTestSuite) Test_CollectDomains_ValidMessage() {
	s.mockEventFetcher.EXPECT().FetchEventLogs(gomock.Any(), s.yahoAddress, string(events.MessageDispatchedSig), big.NewInt(100), big.NewInt(1100)).Return([]types.Log{}, nil)
	s.mockEventFetcher.EXPECT().FetchEventLogs(gomock.Any(), s.yahoAddress, string(events.MessageDispatchedSig), big.NewInt(1101), big.NewInt(2101)).Return([]types.Log{{}}, nil)
	s.mockEventFetcher.EXPECT().FetchEventLogs(gomock.Any(), s.yahoAddress, string(events.MessageDispatchedSig), big.NewInt(2102), big.NewInt(2568)).Return([]types.Log{}, nil)

	domains, err := s.hashiHandler.CollectDomains(big.NewInt(100), big.NewInt(2568))

	s.Nil(err)
	s.Equal(domains, s.domains)
}
