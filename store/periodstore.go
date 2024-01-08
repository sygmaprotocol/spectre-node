// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package store

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/sygmaprotocol/sygma-core/store"
	"github.com/syndtr/goleveldb/leveldb"
)

type PeriodStore struct {
	db store.KeyValueReaderWriter
}

func NewPeriodStore(db store.KeyValueReaderWriter) *PeriodStore {
	return &PeriodStore{
		db: db,
	}
}

// StorePeriod stores latest committee update period per domain
func (ns *PeriodStore) StorePeriod(domainID uint8, period *big.Int) error {
	key := bytes.Buffer{}
	keyS := fmt.Sprintf("chain:%d:period", domainID)
	key.WriteString(keyS)

	err := ns.db.SetByKey(key.Bytes(), period.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// Period queries the blockstore and returns latest period for the requested domain
func (ns *PeriodStore) Period(domainID uint8) (*big.Int, error) {
	key := bytes.Buffer{}
	keyS := fmt.Sprintf("chain:%d:period", domainID)
	key.WriteString(keyS)

	v, err := ns.db.GetByKey(key.Bytes())
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return big.NewInt(0), nil
		}
		return nil, err
	}

	block := big.NewInt(0).SetBytes(v)
	return block, nil
}
