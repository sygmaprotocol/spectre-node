package prover

import "math/big"

type Prover struct{}

func NewProver() *Prover {
	return &Prover{}
}

func (p *Prover) StepProof(epoch *big.Int) ([]byte, error) {
	return []byte{}, nil
}
