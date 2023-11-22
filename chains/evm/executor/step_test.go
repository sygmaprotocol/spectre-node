// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package executor_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/chains/evm/executor"
	"github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/mock"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
	"go.uber.org/mock/gomock"
)

type StepTestSuite struct {
	suite.Suite

	mockStepSubmitter *mock.MockStepSubmitter
	executor          *executor.EVMStepExecutor
}

func TestRunStepTestSuite(t *testing.T) {
	suite.Run(t, new(StepTestSuite))
}

func (s *StepTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockStepSubmitter = mock.NewMockStepSubmitter(ctrl)
	s.executor = executor.NewEVMStepExecutor(1, s.mockStepSubmitter)
}

func (s *StepTestSuite) Test_Execute_SubmissionFails() {
	s.mockStepSubmitter.EXPECT().Step(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))

	err := s.executor.Execute([]*proposal.Proposal{{
		Data: message.StepData{},
	}})

	s.NotNil(err)
}

func (s *StepTestSuite) Test_Execute_Successful() {
	s.mockStepSubmitter.EXPECT().Step(gomock.Any(), gomock.Any(), gomock.Any()).Return(&common.Hash{}, nil)

	err := s.executor.Execute([]*proposal.Proposal{{
		Data: message.StepData{},
	}})

	s.Nil(err)
}
