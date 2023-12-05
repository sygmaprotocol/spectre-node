// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package prover

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

func CountSetBits(arr [64]byte) int {
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
