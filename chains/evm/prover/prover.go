// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package prover

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	ssz "github.com/ferranbt/fastssz"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/spectre-node/chains/evm/message"
	consensus "github.com/umbracle/go-eth-consensus"
)

type StepArgs struct {
	Spec    Spec                                        `json:"spec"`
	Pubkeys [512][48]byte                               `json:"pubkeys"`
	Domain  phase0.Domain                               `json:"domain"`
	Update  *consensus.LightClientFinalityUpdateCapella `json:"light_client_finality_update"`
}

type RotateArgs struct {
	Spec   Spec                                `json:"spec"`
	Update *consensus.LightClientUpdateCapella `json:"light_client_update"`
}

type ProverResponse struct {
	Proof []uint16 `json:"proof"`
}

type CommitmentResponse struct {
	Commitment [32]byte `json:"commitment"`
}

type CommitmentArgs struct {
	Pubkeys [512][48]byte `json:"pubkeys"`
}

type EvmProof[T any] struct {
	Proof []byte
	Input T
}

type LightClient interface {
	FinalityUpdate() (*consensus.LightClientFinalityUpdateCapella, error)
	Updates(period uint64) ([]*consensus.LightClientUpdateCapella, error)
	Bootstrap(blockRoot string) (*consensus.LightClientBootstrapCapella, error)
}

type BeaconClient interface {
	BeaconBlockRoot(ctx context.Context, opts *api.BeaconBlockRootOpts) (*api.Response[*phase0.Root], error)
	Domain(ctx context.Context, domainType phase0.DomainType, epoch phase0.Epoch) (phase0.Domain, error)
}

type ProverClient interface {
	CallFor(ctx context.Context, reply interface{}, method string, args ...interface{}) error
}

type Prover struct {
	lightClient  LightClient
	beaconClient BeaconClient
	proverClient ProverClient

	spec                  Spec
	committeePeriodLength uint64
}

func NewProver(
	proverClient ProverClient,
	beaconClient BeaconClient,
	lightClient LightClient,
	spec Spec,
	committeePeriodLength uint64,
) *Prover {
	return &Prover{
		proverClient:          proverClient,
		spec:                  spec,
		committeePeriodLength: committeePeriodLength,
		beaconClient:          beaconClient,
		lightClient:           lightClient,
	}
}

// StepProof generates the proof for the sync step
func (p *Prover) StepProof() (*EvmProof[message.SyncStepInput], error) {
	args, err := p.stepArgs()
	if err != nil {
		return nil, err
	}

	stepProof, _ := hex.DecodeString("1d563fef05ac85f87d2d233706d9e63424e7e283d0b4862311a6f5730fd46945018d7754b324db35e2e28b885b390ec1eef712129682686cd0f0dbb5a4dff8da140008c2918af1046cdeb4762da6467c972e8399c8b50e5e09741ab285eedde01f15e4ce9ebebfb434b3a6b36dae35b84d9ab522628ebf65732494d507d085922828f65054b6fbc658d2a5451a54acdbc33d7eb409fc3f6f461ff720946e1c6e20d59d8cc788102db2df97919f74356fd5580147fe3177727db6acf71e19271b0722921b83c42eff6fe806b8046562101e55fe0cc390c64d8bb43342c1a6745123f5b9b9afb7301496ebf2b066bba6f3eed61c0ade2ceaa4881a67265dc9651e1505f9ec4b45bea3de08187917b8872d98e2e958fab471f31c11f1cb63c6154a06dc98f32fb1cce16d92117635b2c79b2826debc384003f8d093d4deb0e08a2814ef9c1e0a637806a41be3ecd14ad99a8420c4cb1405381457f93409e04b5fa627cb4722390b320307328a2012c0435f2ded68121e02032b0801421a75d7919123f8360f6aa04a9947ee878ae823cf597f095b893e6d5c9856afc11243b38a5204ff5cf12f14f4b4474eb112e9212f8b324f089928daf1de470f218bb234b7581ee5338c7ed7e1f8c8330065fc7b79543d410a202e0c10e3796b9d76e0c3cb7b16a2935e0959515e0d225b3656da6e0f8c4539da7afc7c6918e0c1c03d84e76913918de9293280c50ba0998f2ab5735995e7d4b53b75ec3f65d6ef0f24567ab11c4237eb0ee9e9868a40094e12d2f33b16917904809ba2610ce5a9a6c2ecf8b70648aab9f3d37062ba4a8d0d24d510b9b397e1e9772f4695a5aad489c2b589da09fbf48829ee1dfc47bd3559c9e97254dca8969517b7fa1e84b80eb99bfc953d1cd41ca05ccc093543d7e1914f6b1bbc86459b64b9ea4e1970ab73976653828108f98f6d73b58639987c4ae3c8ae77ed7607438b9db6560ccd538c9f077b82710a3d9c31efe260a1ecc4c9c2175d818469fd232498f0c79c40f66e25fcb0a59a220f1c8b467bd275617751703b0d05f5399e18250c7925df26e46d4dab24996c1cff9ef5cc03604cf6415cddaa51cbfd67f3bcabe1efc7943c78f3e0504995030392bede71ba9297df982525572e36bb30f7581016e8b53423a62a15853351c1248a3e57e520514849cb4897a5bc0338048dd46f69c852f83bec43d202b419622c1c0cf543b8487b4c526791cde39e9af4d243417e45cfef27c97336a3079d7202fdd2c4be48ea1c6c5554838c5fa0dcf809bba76b575e3f8595939a626588fe15e7241401fd546e7201928479b952d0ed0e5274baf6df971a469822cb7707020562a26aa7d2cac26ed1505ea712e48d1127106792082418b1ff568fb1c0e2a219a38a94ca7c4b171ff6a2e3d0cb98510bfc54940bf3e5ca0ee1e171f20af72c20474a3396bb245e7c661eec4aca2fd6eb94f9208012a66c27aa623560aabd5f2fa863f9a174340d4f72f693a7cc46c5630e388e7295e268ecf1c217701fa11e26be010f91d25418534d11e535e76d30b572afd904fb31644e57886b6132ed0c086f10c86534a257edeafbd29f63dc517453d09f14935feac509f881266dfb690fc35e5c10afe31551772dd8c74e114a71036b0c532d7db4c086618afda53db40cc6c345760cb8a606423871ceed1773356a72b0ff8a1f0ecfb4398111e322c2138df09a2d0fc8ec0a19b184d903b71ab6bf2b4b9cdb28228e7f4ed9d7e38a9e0476e0fa57369ed230e40f258225ef345c5ff4b787494f395c89bf6c246bdec81a89195e5f056466aab781d360438273e3d972a563bd1e70b1841ad380e0db9a286277bceee033848630b0dc3734418568d733bba2b85bebdc9de33a784134d31d875ea6289c250cdc426ccdfabf3b3c0bfb19b425f529b4687f9328ce4c04141964cdcd57e8c4c5d12c34e08a05204183852f55d86e5059016ad135f4a7e3fc238b3284edaba3ffcddf24517ba361a358faf295fbb999459af12418ce4f5c0605e0118f68aa4846c98073495e4f992194e14edaa098987e312f3adca595bb2407ec026724852f2b19a8d16a072e868566d458e9f1eccdb3dab8d60e5b0bc08a173c55c6ed6dae6fe104a7dd55363f82cf57fef440e9989b2b2f2bbcada6717624cc422859ccd5a82c3beda5b9522baf0d33368b9cd9b9a943b9286498b6bd260fa1572622db1f95e1d07a46dc918067d2341101ecef5604fa3824e39c613f581d3ad229d72eb5627cc6b927c83ddb55f837e70a32078d3b9e6778618e1c70ca305acd01f870e1dbb8b30b865b471cc5fdd21966c19318559f83bb9361b7c37d1e3d8e5716c39c5a15e399b214ee543bf0ac1e2e6b6674034ab5619b176e99b202080fb4f174595294615e4e76a94de8ef40356fc716914a46fa25be29291b031c8a93f65bdb862c69c4eb28217b2521bff5737ec27af2fafb27fdcb100076c21c2bf025a6ac84d0dddd87374f505b70059ff25aa4910af4b8b76f3742574d9c1f2fd11e96b334812ce60b1dfb4f851b12c1cced0369a87c7705e8881661962a2f039d4ef2eb76204d3a7dba5051ee3288d49633c66f1448d425be289b1b394717808143cb348ccf70ecfe8d179b228abfc7e47ac856a43e7343a2a06c14d29212fc88c0fbd7333e44094954fd325bdd97964c3777aafb42b60fc671f139642d19556e8b0c1d3c704070c364965abb412b35dd1595bc9963612d6612477fdbc700837b291120bdea7826f176817226bb4671eeaa876ad89bcb8aabdfffa2fe080e6052f9f9e7d07b028fcf4797f82f0f51137f86e2134c23d1a0a7f3018480ba112a58a15ed842af785274c1d3883463d271ec2d1dc6420ffc014791f0ece44d0eedc74397cb15d6beaf22285766b2fff69c35e08838f38b4a4fdef13c5aae43124430124df466d9fa14e955835b2cf10495189a0ca0515ad6c1b5c088c304f018976701fd8d8ec44161e02388ab522fa0e57b46017821bb619300dc67d0b36f1bf300652b0339b0ea2b3b50cf804388c847abc31b3bb29bad467b8ec5c1d85722ce1a8d6c52082844f04c13e28c252415c0c0c0683c11249b4ba10a78dd2e5324b5174ed795f7e92a33eb26198e8e72a57c2af2d151f980daea53e9053b6440056f34cd3a822d90f91064d4648a0d1119ab2eab43da3a9efd698d7fca3478ae100b0462e2cd9cb7241b76b8ee8a5d63d4fe156eeed644622c41932338a1d3082e2244da7a34aa5e7d861d0ce4589f99f55ad42bb40b1de3099000ad59b6722d109c3afbd8a643d8d522ba3e9f4be51ff23e93b999662ef1142d4d946395d12007226d919044c38d61a2336649de78bac2ff0d162e72a20eef9fea128114aa13060ad83c34aef8e4e6378559aaa3a715ce101bc3410d0220a14e4fcda043d5e7023619538c22c15f05c547e942364888701acf9d4efbba39e72a8d60fd9f1c3520973c6dbe009cb7bf912e1f86be3e28e49c1e6e3f75a58262a8f9ea66f10cd828bd9346da7008a0c68c3c0385f04121bf91d7765a9583d150400d5cd337239d28bd9346da7008a0c68c3c0385f04121bf91d7765a9583d150400d5cd337239d1c2fe8c9209e99a1749e3cac55fb99c25a7fe16c8b54ac4469cfdb74c8c37d6f02d382dc5f6fa15f452c5154b7ec776bbffe69a95aea4832428fc428257f27661a928315abd0a9f2b536ada8d727d422ec81214fff32943b72d8a80c0f19a7d92f56d2672e4c9d49bad3c516c1aa585a6c97df310b4547ac763e94a49d43c2cb0ddc5e53bf87bf135e8790373bfa466e76dce796d2b4556c4ea2e2accf493d7622255a6aafcf5f6462e8205fef26ee67aa64dac99a819c320d593ce1f9cb7d76291177b2fecd2013e8b36a4016c6ec942150e1aec3e7ee0f3adb93ebe2d3b3721c9582c990beebe8f7f8eed6092b7a77619eea05d76bd9cc525db15721982c791cf34a7503f71e093f1f7d638f62601b2b7d1c5528c3e8d39544afc53282d1a71ea4c2753891349ce7449e611b412b2c4420b927a911e939fdf507fa54f340d014d37c756726e95e93a5c89d5f700b75619e83042ae55983ef1c07af32ce8a57058bb8305fb54f290d465a234d4521f1b38f122b3ecb6bf6637beef6e7ec60711a46f078ecdf0c759d01d368637673a6126e4306c64f23aa22839ba6f32acb0d1e1233d53550b022506eb27921180a1541cad019107da674b3be3fb09b24e5ea17bc64f5f8f5d10f36c87e19991b4dcad2fcab3120022e423864f2f0e11e2cee1d07accbe92c12fde746f03bd2fb58d8da8213c433c2d8a1b634859d86ff62811a1569c46c8bb092186c18a4e13a701dc68b97a5e516870682c0606a9a0c977607610d42099c69aeaf3f4b534b10462d101462fba115d7274e225bccfd02a2af09e108c5fbeee4ddc731ed08c00164f511c33a61fa3994d69976f507978c809a006f15059abf42b542c8086c127cd601666f971b20c6049f49b828af0c3fce0f09df6f9ca882bc806f074d0b99f4094d873742b3c373dd5c8408653ee7eab437064be30171650e6aefa92a9ac67e153988a6ce1e051440c9af4082a8cb62414510bda393cd0966c2a673657e03c511449ddd103bf9a9be8ac4850174e3ff5a8d2fffdc56ab27c17914d8411da73a9dabd81dd56671c80b19a605a2e86c1233492b5e15c3cd582a46beae17252802cb5f75df026d67a095ae70dc359157f254970deae0e3c9b25ac75dbeca648e57c113e8d6d6ffa2afa23459afef729b1f233120c7c9c0fe9b2fd7d3dd8c00a88e4fc038233e02a97f5cd12f0af58554afd14a21afa82ddad5b28637f7effc0d426ab96d39bd9118d464aef59d1f529e513a971972c74663d1c6f635a47b408138008d6febfd8ce17a2494af6547befd8f819627693a28070912e4eb33351c53b90a270db16972a9dc03810fbc909bf4f8305704385f17fc38e78b1641b3b7086a039a8742f04ed4b6f65ce7557b50b41ad3442367d685dca260ab7ee3e75bc1ca0ab8371903a61c33381ecb391f4a06b4de201a4d49c01c09827456c58aac2afe30433836059dce6bd74e391e9daa478a19850c8fd3e2e67b468553d1bee531dfa76e0d5a4a37f29bc968568dd1d02e096510165eabcb573bb11cf243b28d91785e35cadf9ec0e94237eb68ef472d3fe45fd804d26b5d0b20511eb39fb22e0865bc408bfcbd1ea8bb200393ac1097c5b2a97201e8a41212737f1e75b7996ce353280f3447049db8650f3dafbb755d09a7a0b12961a85cdda6a611c493cedaf29f5ebabbe7e2013183de925d095109ef71ed5f11f841299ad3ba17e7709df29aded571756301a23bc10240abc6428ff53cef5704dabc42e76af9eb024692125030b75d106338f27600f7ee7cd54a5e2a3d52c40188392e3b78b1615ac7246fe4552c11383210cfc805be639e896378fd7201f41b75ad318d105664c31d593b97030acbe11a6aeb0300aea88e21b9953e1675f81a48b13adbfa074c38b1472dd87e95060ef7a6ba71ac8cfb2733511b5261da57151a68a4f323333b62667d938b49448c3c1192dd9c2760415cd95cf56730d9330e7a722fdfe350c30a1d8ecf099edd10c6124c16a8050e9adee35567ab6f535e1eb367162191e58b4b7e1d2efed6cf67cb78f29de61197015daf44d7d00e0cee18700d7db2b4f99bade4f11a61a6727b554c3ff38f7e8347c619b99d6712dbe2224b0a61ddf3771cb3c2c000dd1697c631d8fec8715df3778962a662215ea0fd09db8b65ca4c2c3ff5ba19002cc06524581256d5928c5b7451279fb0723bc010154b6c55282ab353d4dcd1cb5fea58b250d116b41dbb4d5a85f6589d490a1b7a")
	/*
		updateSzz, _ := args.Update.MarshalSSZ()
		pubkeysSZZ := make([]byte, 0)
		for i := 0; i < 512; i++ {
			pubkeysSZZ = append(pubkeysSZZ, args.Pubkeys[i][:]...)
		}

		type stepArgs struct {
			Spec    Spec     `json:"spec"`
			Pubkeys []uint16 `json:"pubkeys"`
			Domain  []uint16 `json:"domain"`
			Update  []uint16 `json:"light_client_finality_update"`
		}
		var resp ProverResponse
		err = p.proverClient.CallFor(context.Background(), &resp, "genEvmProofAndInstancesStepSyncCircuitWithWitness", stepArgs{
			Spec:    args.Spec,
			Pubkeys: ByteArrayToU16Array(pubkeysSZZ),
			Update:  ByteArrayToU16Array(updateSzz),
			Domain:  ByteArrayToU16Array(args.Domain[:]),
		})
		if err != nil {
			return nil, err
		}

		log.Info().Msgf("Generated step proof %s", hex.EncodeToString(U16ArrayToByteArray(resp.Proof)))
	*/

	finalizedHeaderRoot, err := args.Update.FinalizedHeader.Header.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	executionRoot, err := args.Update.FinalizedHeader.Execution.HashTreeRoot()
	if err != nil {
		return nil, err
	}

	proof := &EvmProof[message.SyncStepInput]{
		Proof: stepProof,
		Input: message.SyncStepInput{
			AttestedSlot:         args.Update.AttestedHeader.Header.Slot,
			FinalizedSlot:        args.Update.FinalizedHeader.Header.Slot,
			Participation:        uint64(countSetBits(args.Update.SyncAggregate.SyncCommiteeBits)),
			FinalizedHeaderRoot:  finalizedHeaderRoot,
			ExecutionPayloadRoot: executionRoot,
		},
	}
	return proof, nil
}

// RotateProof generates the proof for the sync committee rotation for the period
func (p *Prover) RotateProof(epoch uint64) (*EvmProof[message.RotateInput], error) {
	args, err := p.rotateArgs(epoch)
	if err != nil {
		return nil, err
	}

	syncCommiteeRoot, err := p.committeeKeysRoot(args.Update.NextSyncCommittee.PubKeys)
	if err != nil {
		return nil, err
	}

	var commitmentResp CommitmentResponse
	err = p.proverClient.CallFor(context.Background(), &commitmentResp, "syncCommitteePoseidonCompressed", CommitmentArgs{Pubkeys: args.Update.NextSyncCommittee.PubKeys})
	if err != nil {
		return nil, err
	}

	updateSzz, _ := args.Update.MarshalSSZ()

	type rotateArgs struct {
		Update []uint16 `json:"light_client_update"`
		Spec   Spec     `json:"spec"`
	}
	var resp ProverResponse
	err = p.proverClient.CallFor(context.Background(), &resp, "genEvmProofAndInstancesRotationCircuitWithWitness", rotateArgs{Update: ByteArrayToU16Array(updateSzz), Spec: args.Spec})
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Generated rotate proof %s", hex.EncodeToString(U16ArrayToByteArray(resp.Proof)))

	proof := &EvmProof[message.RotateInput]{
		Proof: U16ArrayToByteArray(resp.Proof[:]),
		Input: message.RotateInput{
			SyncCommitteeSSZ:      syncCommiteeRoot,
			SyncCommitteePoseidon: commitmentResp.Commitment,
		},
	}
	return proof, nil
}

func (p *Prover) stepArgs() (*StepArgs, error) {
	update, err := p.lightClient.FinalityUpdate()
	if err != nil {
		return nil, err
	}
	blockRoot, err := p.beaconClient.BeaconBlockRoot(context.Background(), &api.BeaconBlockRootOpts{
		Block: fmt.Sprint(update.FinalizedHeader.Header.Slot),
	})
	if err != nil {
		return nil, err
	}
	bootstrap, err := p.lightClient.Bootstrap(blockRoot.Data.String())
	if err != nil {
		return nil, err
	}
	pubkeys := bootstrap.CurrentSyncCommittee.PubKeys

	domain, err := p.beaconClient.Domain(context.Background(), SYNC_COMMITTEE_DOMAIN, phase0.Epoch(update.FinalizedHeader.Header.Slot/32))
	if err != nil {
		return nil, err
	}

	return &StepArgs{
		Pubkeys: pubkeys,
		Domain:  domain,
		Update:  update,
		Spec:    p.spec,
	}, nil
}

func (p *Prover) rotateArgs(epoch uint64) (*RotateArgs, error) {
	period := epoch / p.committeePeriodLength
	updates, err := p.lightClient.Updates(period)
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return nil, fmt.Errorf("missing light client updates")
	}
	update := updates[0]
	return &RotateArgs{
		Update: update,
		Spec:   p.spec,
	}, nil
}

func (p *Prover) committeeKeysRoot(pubkeys [512][48]byte) ([32]byte, error) {
	keysSSZ := make([]byte, 0)
	for i := 0; i < 512; i++ {
		keysSSZ = append(keysSSZ, pubkeys[i][:]...)
	}

	hh := ssz.NewHasher()
	hh.PutBytes(keysSSZ)
	return hh.HashRoot()
}

func ByteArrayToU16Array(src []byte) []uint16 {
	dst := make([]uint16, len(src))
	for i, value := range src {
		dst[i] = uint16(value)
	}
	return dst
}

func U16ArrayTo32ByteArray(src []uint16) [32]byte {
	dst := [32]byte{}
	for i, value := range src {
		dst[i] = byte(value)
	}
	return dst
}

func U16ArrayToByteArray(src []uint16) []byte {
	dst := make([]byte, len(src))
	for i, value := range src {
		dst[i] = byte(value)
	}
	return dst
}
func countSetBits(arr [64]byte) int {
	count := 0

	for _, b := range arr {
		for i := 0; i < 8; i++ {
			// Check if the i-th bit is set (1)
			if b&(1<<i) != 0 {
				count++
			}
		}
	}

	return count
}
