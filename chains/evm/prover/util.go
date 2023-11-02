package prover

const EPOCH_TIME = 432000
const UNIX_BEGINNING = 1506203091

func EpochFromTimestamp(timestamp uint64) uint64 {
	return (timestamp - UNIX_BEGINNING) / EPOCH_TIME
}
