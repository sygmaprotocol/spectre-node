// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package abi

const SpectreABI = `[
  {
    "inputs": [
      {
        "internalType": "uint8[]",
        "name": "domainIDS",
        "type": "uint8[]"
      },
      {
        "internalType": "address[]",
        "name": "spectreAddresses",
        "type": "address[]"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "constructor"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": false,
        "internalType": "uint8",
        "name": "sourceDomainID",
        "type": "uint8"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "slot",
        "type": "uint256"
      }
    ],
    "name": "CommitteeRotated",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "account",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "sender",
        "type": "address"
      }
    ],
    "name": "RoleGranted",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "account",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "sender",
        "type": "address"
      }
    ],
    "name": "RoleRevoked",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": false,
        "internalType": "uint8",
        "name": "sourceDomainID",
        "type": "uint8"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "slot",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "bytes32",
        "name": "stateRoot",
        "type": "bytes32"
      }
    ],
    "name": "StateRootSubmitted",
    "type": "event"
  },
  {
    "inputs": [],
    "name": "DEFAULT_ADMIN_ROLE",
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
    "name": "STATE_ROOT_INDEX",
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
        "internalType": "uint8",
        "name": "sourceDomainID",
        "type": "uint8"
      },
      {
        "internalType": "address",
        "name": "spectreAddress",
        "type": "address"
      }
    ],
    "name": "adminSetSpectreAddress",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      }
    ],
    "name": "getRoleAdmin",
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
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      },
      {
        "internalType": "uint256",
        "name": "index",
        "type": "uint256"
      }
    ],
    "name": "getRoleMember",
    "outputs": [
      {
        "internalType": "address",
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
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      }
    ],
    "name": "getRoleMemberCount",
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
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      },
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "getRoleMemberIndex",
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
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      },
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "grantRole",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      },
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "hasRole",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      },
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "renounceRole",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "role",
        "type": "bytes32"
      },
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "revokeRole",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint8",
        "name": "sourceDomainID",
        "type": "uint8"
      },
      {
        "components": [
          {
            "internalType": "bytes32",
            "name": "syncCommitteeSSZ",
            "type": "bytes32"
          },
          {
            "internalType": "uint256",
            "name": "syncCommitteePoseidon",
            "type": "uint256"
          },
          {
            "internalType": "uint256[12]",
            "name": "accumulator",
            "type": "uint256[12]"
          }
        ],
        "internalType": "struct ISpectre.RotateInput",
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
          },
          {
            "internalType": "uint256[12]",
            "name": "accumulator",
            "type": "uint256[12]"
          }
        ],
        "internalType": "struct ISpectre.SyncStepInput",
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
        "internalType": "uint8",
        "name": "",
        "type": "uint8"
      }
    ],
    "name": "spectreContracts",
    "outputs": [
      {
        "internalType": "address",
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
        "internalType": "uint8",
        "name": "",
        "type": "uint8"
      },
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "name": "stateRoots",
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
    "inputs": [
      {
        "internalType": "uint8",
        "name": "sourceDomainID",
        "type": "uint8"
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
          },
          {
            "internalType": "uint256[12]",
            "name": "accumulator",
            "type": "uint256[12]"
          }
        ],
        "internalType": "struct ISpectre.SyncStepInput",
        "name": "input",
        "type": "tuple"
      },
      {
        "internalType": "bytes",
        "name": "stepProof",
        "type": "bytes"
      },
      {
        "internalType": "bytes32",
        "name": "stateRoot",
        "type": "bytes32"
      },
      {
        "internalType": "bytes[]",
        "name": "stateRootProof",
        "type": "bytes[]"
      }
    ],
    "name": "step",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]`
