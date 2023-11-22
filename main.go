// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"fmt"
	"math/big"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/attestantio/go-eth2-client/http"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	evmConfig "github.com/sygmaprotocol/spectre-node/chains/evm/config"
	"github.com/sygmaprotocol/spectre-node/chains/evm/contracts"
	"github.com/sygmaprotocol/spectre-node/chains/evm/executor"
	"github.com/sygmaprotocol/spectre-node/chains/evm/lightclient"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/events/handlers"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/spectre-node/config"
	"github.com/sygmaprotocol/sygma-core/chains/evm"
	"github.com/sygmaprotocol/sygma-core/chains/evm/client"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor/gas"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor/monitored"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor/transaction"
	"github.com/sygmaprotocol/sygma-core/crypto/secp256k1"
	"github.com/sygmaprotocol/sygma-core/observability"
	"github.com/sygmaprotocol/sygma-core/relayer"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logLevel, err := zerolog.ParseLevel(cfg.Observability.LogLevel)
	if err != nil {
		panic(err)
	}
	observability.ConfigureLogger(logLevel, os.Stdout)

	log.Info().Msg("Loaded configuration")

	domains := make([]uint8, 0)
	for domain := range cfg.Domains {
		domains = append(domains, domain)
	}

	msgChan := make(chan []*message.Message)
	chains := make(map[uint8]relayer.RelayedChain)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for id, nType := range cfg.Domains {
		switch nType {
		case "evm":
			{
				config, err := evmConfig.LoadEVMConfig(id)
				if err != nil {
					panic(err)
				}

				kp, err := secp256k1.NewKeypairFromString(config.Key)
				if err != nil {
					panic(err)
				}

				client, err := client.NewEVMClient(config.Endpoint, kp)
				if err != nil {
					panic(err)
				}

				gasPricer := gas.NewLondonGasPriceClient(client, &gas.GasPricerOpts{
					UpperLimitFeePerGas: big.NewInt(config.MaxGasPrice),
					GasPriceFactor:      big.NewFloat(config.GasMultiplier),
				})
				t := monitored.NewMonitoredTransactor(transaction.NewTransaction, gasPricer, client, big.NewInt(config.MaxGasPrice), big.NewInt(config.GasIncreasePercentage))
				go t.Monitor(ctx, time.Minute*3, time.Minute*10, time.Minute)

				beaconClient, err := http.New(ctx,
					http.WithAddress(config.BeaconEndpoint),
					http.WithLogLevel(logLevel),
					http.WithTimeout(time.Second*30),
				)
				if err != nil {
					panic(err)
				}
				beaconProvider := beaconClient.(*http.Service)

				proverClient, err := rpc.DialHTTP("tcp", cfg.Prover.URL)
				if err != nil {
					panic(err)
				}
				lightClient := lightclient.NewLightClient(config.BeaconEndpoint)
				p := prover.NewProver(proverClient, beaconProvider, lightClient, prover.Spec(config.Spec), config.CommitteePeriodLength)
				routerAddress := common.HexToAddress(config.Router)
				depositHandler := handlers.NewDepositEventHandler(msgChan, client, beaconProvider, p, routerAddress, id, config.BlockInterval)
				rotateHandler := handlers.NewRotateHandler(msgChan, beaconProvider, p, id, domains)
				listener := listener.NewEVMListener(beaconProvider, []listener.EventHandler{rotateHandler, depositHandler}, id, time.Duration(config.RetryInterval)*time.Second)

				messageHandler := message.NewMessageHandler()

				rotateMessageHandler := evmMessage.EvmRotateHandler{}
				stepMessageHandler := evmMessage.EvmStepHandler{}
				messageHandler.RegisterMessageHandler(evmMessage.EVMRotateMessage, &rotateMessageHandler)
				messageHandler.RegisterMessageHandler(evmMessage.EVMStepMessage, &stepMessageHandler)

				spectre := contracts.NewSpectreContract(common.HexToAddress(config.Spectre), t)
				executor := executor.NewEVMExecutor(id, spectre)

				chain := evm.NewEVMChain(listener, messageHandler, executor, id, nil)
				chains[id] = chain
			}
		default:
			{
				panic(fmt.Sprintf("invalid network type %s for id %d", nType, id))
			}
		}
	}

	r := relayer.NewRelayer(chains)
	go r.Start(ctx, msgChan)

	sysErr := make(chan os.Signal, 1)
	signal.Notify(sysErr,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGQUIT)
	log.Info().Msgf("Started spectre node")

	se := <-sysErr
	log.Info().Msgf("terminating got ` [%v] signal", se)
}
