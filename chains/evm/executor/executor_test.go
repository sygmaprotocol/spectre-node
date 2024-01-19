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

type ExecutorTestSuite struct {
	suite.Suite

	mockProofSubmitter *mock.MockProofSubmitter
	executor           *executor.EVMExecutor
}

func TestRunStepTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}

func (s *ExecutorTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockProofSubmitter = mock.NewMockProofSubmitter(ctrl)
	s.executor = executor.NewEVMExecutor(1, s.mockProofSubmitter)
}

func (s *ExecutorTestSuite) Test_Execute_InvalidPropType() {
	err := s.executor.Execute([]*proposal.Proposal{{
		Type: "invalid",
	}})

	s.NotNil(err)
}

func (s *ExecutorTestSuite) Test_Execute_Step_SubmissionFails() {
	s.mockProofSubmitter.EXPECT().Step(uint8(1), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))

	err := s.executor.Execute([]*proposal.Proposal{{
		Data:   message.StepData{},
		Type:   message.EVMStepProposal,
		Source: 1,
	}})

	s.NotNil(err)
}

func (s *ExecutorTestSuite) Test_Execute_Step_Successful() {
	s.mockProofSubmitter.EXPECT().Step(uint8(1), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&common.Hash{}, nil)

	err := s.executor.Execute([]*proposal.Proposal{{
		Data:   message.StepData{},
		Type:   message.EVMStepProposal,
		Source: 1,
	}})

	s.Nil(err)
}

func (s *ExecutorTestSuite) Test_Execute_Rotate_SubmissionFails() {
	s.mockProofSubmitter.EXPECT().Rotate(uint8(1), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))

	err := s.executor.Execute([]*proposal.Proposal{{
		Data:   message.RotateData{},
		Type:   message.EVMRotateProposal,
		Source: 1,
	}})

	s.NotNil(err)
}

func (s *ExecutorTestSuite) Test_Execute_Rotate_Successful() {
	s.mockProofSubmitter.EXPECT().Rotate(uint8(1), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&common.Hash{}, nil)

	err := s.executor.Execute([]*proposal.Proposal{{
		Data:   message.RotateData{},
		Type:   message.EVMRotateProposal,
		Source: 1,
	}})

	s.Nil(err)
}
