// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package store_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/mock"
	"github.com/sygmaprotocol/spectre-node/store"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/mock/gomock"
)

type PeriodStoreTestSuite struct {
	suite.Suite
	periodStore          *store.PeriodStore
	keyValueReaderWriter *mock.MockKeyValueReaderWriter
}

func TestRunPeriodStoreTestSuite(t *testing.T) {
	suite.Run(t, new(PeriodStoreTestSuite))
}

func (s *PeriodStoreTestSuite) SetupTest() {
	gomockController := gomock.NewController(s.T())
	s.keyValueReaderWriter = mock.NewMockKeyValueReaderWriter(gomockController)
	s.periodStore = store.NewPeriodStore(s.keyValueReaderWriter)
}

func (s *PeriodStoreTestSuite) Test_StorePeriod_FailedStore() {
	key := "chain:1:period"
	s.keyValueReaderWriter.EXPECT().SetByKey([]byte(key), []byte{5}).Return(errors.New("error"))

	err := s.periodStore.StorePeriod(1, big.NewInt(5))

	s.NotNil(err)
}

func (s *PeriodStoreTestSuite) Test_StorePeriod_SuccessfulStore() {
	key := "chain:1:period"
	s.keyValueReaderWriter.EXPECT().SetByKey([]byte(key), []byte{5}).Return(nil)

	err := s.periodStore.StorePeriod(1, big.NewInt(5))

	s.Nil(err)
}

func (s *PeriodStoreTestSuite) Test_Period_FailedFetch() {
	key := "chain:1:period"
	s.keyValueReaderWriter.EXPECT().GetByKey([]byte(key)).Return(nil, errors.New("error"))

	_, err := s.periodStore.Period(1)

	s.NotNil(err)
}

func (s *PeriodStoreTestSuite) TestGetNonce_NonceNotFound() {
	key := "chain:1:period"
	s.keyValueReaderWriter.EXPECT().GetByKey([]byte(key)).Return(nil, leveldb.ErrNotFound)

	period, err := s.periodStore.Period(1)

	s.Nil(err)
	s.Equal(period, big.NewInt(0))
}

func (s *PeriodStoreTestSuite) TestGetNonce_SuccessfulFetch() {
	key := "chain:1:period"
	s.keyValueReaderWriter.EXPECT().GetByKey([]byte(key)).Return([]byte{5}, nil)

	period, err := s.periodStore.Period(1)

	s.Nil(err)
	s.Equal(period, big.NewInt(5))
}
