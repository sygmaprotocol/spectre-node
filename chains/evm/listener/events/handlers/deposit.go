// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"math/big"
	"strings"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/spectre-node/chains/evm/abi"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
)

type EventFetcher interface {
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
}

type Prover interface {
	StepProof() (*prover.EvmProof[evmMessage.SyncStepInput], error)
	RotateProof(epoch uint64) (*prover.EvmProof[evmMessage.RotateInput], error)
}

type BlockFetcher interface {
	SignedBeaconBlock(ctx context.Context, opts *api.SignedBeaconBlockOpts) (*api.Response[*spec.VersionedSignedBeaconBlock], error)
}

type DepositEventHandler struct {
	msgChan chan []*message.Message

	eventFetcher EventFetcher
	blockFetcher BlockFetcher
	prover       Prover

	domainID      uint8
	domains       []uint8
	blockInterval uint64
	routerABI     ethereumABI.ABI
	routerAddress common.Address
}

func NewDepositEventHandler(
	msgChan chan []*message.Message,
	eventFetcher EventFetcher,
	blockFetcher BlockFetcher,
	prover Prover,
	routerAddress common.Address,
	domainID uint8,
	domains []uint8,
	blockInterval uint64,
) *DepositEventHandler {
	routerABI, _ := ethereumABI.JSON(strings.NewReader(abi.RouterABI))
	return &DepositEventHandler{
		eventFetcher:  eventFetcher,
		blockFetcher:  blockFetcher,
		prover:        prover,
		domains:       domains,
		routerAddress: routerAddress,
		routerABI:     routerABI,
		msgChan:       msgChan,
		domainID:      domainID,
		blockInterval: blockInterval,
	}
}

// HandleEvents fetches deposit events and if deposits exists, submits a step message
// to be executed on the destination network
func (h *DepositEventHandler) HandleEvents(checkpoint *apiv1.Finality) error {
	/*
			startBlock, endBlock, err := h.blockrange(checkpoint)
			if err != nil {
				return err
			}

			deposits, err := h.fetchDeposits(startBlock, endBlock)
			if err != nil {
				return fmt.Errorf("unable to fetch deposit events because of: %+v", err)
			}
			domainDeposits := make(map[uint8][]*events.Deposit)
			for _, d := range deposits {
				domainDeposits[d.DestinationDomainID] = append(domainDeposits[d.DestinationDomainID], d)
			}
			if len(domainDeposits) == 0 {
				return nil
			}

		log.Info().Uint8("domainID", h.domainID).Msgf("Found deposits between blocks %s-%s", startBlock, endBlock)

	*/

	proof, err := h.prover.StepProof()
	if err != nil {
		return err
	}
	for _, destDomain := range h.domains {
		log.Debug().Uint8("domainID", h.domainID).Msgf("Sending step message to domain %d", destDomain)
		h.msgChan <- []*message.Message{
			evmMessage.NewEvmStepMessage(
				h.domainID,
				destDomain,
				evmMessage.StepData{
					Proof: proof.Proof,
					Args:  proof.Input,
				},
			),
		}
	}
	return nil
}

/*
func (h *DepositEventHandler) blockrange(checkpoint *apiv1.Finality) (*big.Int, *big.Int, error) {
	justifiedRoot, err := h.blockFetcher.SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: checkpoint.Justified.Root.String(),
	})
	if err != nil {
		return nil, nil, err
	}

	endBlock := big.NewInt(int64(justifiedRoot.Data.Capella.Message.Body.ExecutionPayload.BlockNumber))
	startBlock := new(big.Int).Sub(endBlock, big.NewInt(int64(h.blockInterval)))
	return startBlock, endBlock, nil
}

func (h *DepositEventHandler) fetchDeposits(startBlock *big.Int, endBlock *big.Int) ([]*events.Deposit, error) {
	logs, err := h.eventFetcher.FetchEventLogs(context.Background(), h.routerAddress, string(events.DepositSig), startBlock, endBlock)
	if err != nil {
		return nil, err
	}

	deposits := make([]*events.Deposit, 0)
	for _, dl := range logs {
		d, err := h.unpackDeposit(dl.Data)
		if err != nil {
			log.Error().Msgf("Failed unpacking deposit event log: %v", err)
			continue
		}
		d.SenderAddress = common.BytesToAddress(dl.Topics[1].Bytes())

		log.Debug().Msgf("Found deposit log in block: %d, TxHash: %s, contractAddress: %s, sender: %s", dl.BlockNumber, dl.TxHash, dl.Address, d.SenderAddress)
		deposits = append(deposits, d)
	}

	return deposits, nil
}

func (h *DepositEventHandler) unpackDeposit(data []byte) (*events.Deposit, error) {
	var d events.Deposit
	err := h.routerABI.UnpackIntoInterface(&d, "Deposit", data)
	if err != nil {
		return &events.Deposit{}, err
	}

	return &d, nil
}
*/
