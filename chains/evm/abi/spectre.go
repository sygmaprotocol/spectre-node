package abi

const SpectreABI = `[
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "_stepVerifierAddress",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "_committeeUpdateVerifierAddress",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "_initialSyncPeriod",
				"type": "uint256"
			},
			{
				"internalType": "bytes32",
				"name": "_initialSyncCommitteePoseidon",
				"type": "bytes32"
			},
			{
				"internalType": "uint256",
				"name": "_slotsPerPeriod",
				"type": "uint256"
			}
		],
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"name": "blockHeaderRoots",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "committeeUpdateVerifier",
		"outputs": [
			{
				"internalType": "contract CommitteeUpdateVerifier",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"name": "executionStateRoots",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "head",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "bytes32",
						"name": "syncCommitteeSSZ",
						"type": "bytes32"
					},
					{
						"internalType": "bytes32",
						"name": "syncCommitteePoseidon",
						"type": "bytes32"
					}
				],
				"internalType": "struct RotateLib.RotateInput",
				"name": "rotateInput",
				"type": "tuple"
			},
			{
				"internalType": "bytes",
				"name": "rotateProof",
				"type": "bytes"
			},
			{
				"components": [
					{
						"internalType": "uint64",
						"name": "attestedSlot",
						"type": "uint64"
					},
					{
						"internalType": "uint64",
						"name": "finalizedSlot",
						"type": "uint64"
					},
					{
						"internalType": "uint64",
						"name": "participation",
						"type": "uint64"
					},
					{
						"internalType": "bytes32",
						"name": "finalizedHeaderRoot",
						"type": "bytes32"
					},
					{
						"internalType": "bytes32",
						"name": "executionPayloadRoot",
						"type": "bytes32"
					}
				],
				"internalType": "struct SyncStepLib.SyncStepInput",
				"name": "stepInput",
				"type": "tuple"
			},
			{
				"internalType": "bytes",
				"name": "stepProof",
				"type": "bytes"
			}
		],
		"name": "rotate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "uint64",
						"name": "attestedSlot",
						"type": "uint64"
					},
					{
						"internalType": "uint64",
						"name": "finalizedSlot",
						"type": "uint64"
					},
					{
						"internalType": "uint64",
						"name": "participation",
						"type": "uint64"
					},
					{
						"internalType": "bytes32",
						"name": "finalizedHeaderRoot",
						"type": "bytes32"
					},
					{
						"internalType": "bytes32",
						"name": "executionPayloadRoot",
						"type": "bytes32"
					}
				],
				"internalType": "struct SyncStepLib.SyncStepInput",
				"name": "input",
				"type": "tuple"
			},
			{
				"internalType": "bytes",
				"name": "proof",
				"type": "bytes"
			}
		],
		"name": "step",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "stepVerifier",
		"outputs": [
			{
				"internalType": "contract SyncStepVerifier",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"name": "syncCommitteePoseidons",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`
