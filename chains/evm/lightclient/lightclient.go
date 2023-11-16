// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package lightclient

import (
	"fmt"
	"io"
	"net/http"

	consensus "github.com/umbracle/go-eth-consensus"
	encoding "github.com/umbracle/go-eth-consensus/http"
)

type response struct {
	Data *consensus.LightClientFinalityUpdateCapella `json:"data,omitempty"`
}

type LightClient struct {
	beaconURL string
}

func NewLightClient(url string) *LightClient {
	return &LightClient{
		beaconURL: url,
	}
}

// Updates fetches light client updates for sync committee period
func (c *LightClient) Updates(period uint64) ([]*consensus.LightClientUpdateCapella, error) {
	resp, err := http.Get(fmt.Sprintf("%s/eth/v1/beacon/light_client/updates?start_period=%d&count=1", c.beaconURL, period))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type response struct {
		Data *consensus.LightClientUpdateCapella `json:"data"`
	}
	apiResponse := make([]response, 0)
	if err := c.decodeResp(resp, &apiResponse); err != nil {
		return nil, err
	}

	updates := make([]*consensus.LightClientUpdateCapella, len(apiResponse))
	for i, update := range apiResponse {
		updates[i] = update.Data
	}

	return updates, err
}

// FinalityUpdate returns the latest finalized light client update
func (c *LightClient) FinalityUpdate() (*consensus.LightClientFinalityUpdateCapella, error) {
	resp, err := http.Get(fmt.Sprintf("%s/eth/v1/beacon/light_client/finality_update", c.beaconURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type response struct {
		Data *consensus.LightClientFinalityUpdateCapella `json:"data,omitempty"`
	}
	var apiResponse response
	if err := c.decodeResp(resp, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, err
}

// Boostrap returns the latest light client bootstrap for the given block root
func (c *LightClient) Bootstrap(blockRoot string) (*consensus.LightClientBootstrapCapella, error) {
	resp, err := http.Get(fmt.Sprintf("%s/eth/v1/beacon/light_client/boostrap/%s", c.beaconURL, blockRoot))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type response struct {
		Data *consensus.LightClientBootstrapCapella `json:"data,omitempty"`
	}
	var apiResponse response
	if err := c.decodeResp(resp, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, err
}

func (c *LightClient) decodeResp(resp *http.Response, out interface{}) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed fetching light client data with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return encoding.Unmarshal(data, &out, false)
}
