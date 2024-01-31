// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"fmt"
	"math/big"
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
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/handlers"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/spectre-node/config"
	"github.com/sygmaprotocol/spectre-node/health"
	"github.com/sygmaprotocol/spectre-node/store"
	"github.com/sygmaprotocol/sygma-core/chains/evm"
	"github.com/sygmaprotocol/sygma-core/chains/evm/client"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor/gas"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor/monitored"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor/transaction"
	"github.com/sygmaprotocol/sygma-core/crypto/secp256k1"
	"github.com/sygmaprotocol/sygma-core/observability"
	"github.com/sygmaprotocol/sygma-core/relayer"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-core/store/lvldb"
	"github.com/ybbus/jsonrpc/v3"
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

	go health.StartHealthEndpoint(cfg.Observability.HealthPort)

	var db *lvldb.LVLDB
	for {
		db, err = lvldb.NewLvlDB(cfg.Store.Path)
		if err != nil {
			log.Error().Err(err).Msg("Unable to connect to blockstore file, retry in 10 seconds")
			time.Sleep(10 * time.Second)
		} else {
			log.Info().Msg("Successfully connected to blockstore file")
			break
		}
	}
	periodStore := store.NewPeriodStore(db)

	domains := make([]uint8, 0)
	for domain := range cfg.Domains {
		domains = append(domains, domain)
	}

	proverClient := jsonrpc.NewClient(cfg.Prover.URL)

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

				storedPeriod, err := periodStore.Period(id)
				if err != nil {
					panic(err)
				}
				var latestPeriod *big.Int
				if (storedPeriod.Uint64() >= config.StartingPeriod) && !config.ForcePeriod {
					latestPeriod = storedPeriod
				} else {
					latestPeriod = big.NewInt(int64(config.StartingPeriod))
				}

				lightClient := lightclient.NewLightClient(config.BeaconEndpoint)
				p := prover.NewProver(proverClient, beaconProvider, lightClient, prover.Spec(config.Spec))
				routerAddress := common.HexToAddress(config.Router)
				stepHandler := handlers.NewStepEventHandler(msgChan, client, beaconProvider, p, routerAddress, id, domains, config.BlockInterval)
				rotateHandler := handlers.NewRotateHandler(msgChan, periodStore, p, id, domains, config.CommitteePeriodLength, latestPeriod)
				listener := listener.NewEVMListener(beaconProvider, []listener.EventHandler{rotateHandler, stepHandler}, id, time.Duration(config.RetryInterval)*time.Second)

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
