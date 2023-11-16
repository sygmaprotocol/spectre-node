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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	evmConfig "github.com/sygmaprotocol/spectre-node/chains/evm/config"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener"
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

	msgChan := make(chan []*message.Message)
	chains := make(map[uint8]relayer.RelayedChain)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for id, nType := range cfg.Chains {
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
				)
				if err != nil {
					panic(err)
				}
				beaconProvider := beaconClient.(*http.Service)
				listener := listener.NewEVMListener(beaconProvider, []listener.EventHandler{}, id, time.Duration(config.RetryInterval)*time.Second)

				chain := evm.NewEVMChain(listener, nil, nil, id, nil)
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

	select {
	case se := <-sysErr:
		log.Info().Msgf("terminating got ` [%v] signal", se)
		return
	}
}
