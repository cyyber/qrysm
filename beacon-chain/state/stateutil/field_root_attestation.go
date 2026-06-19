package stateutil

// RootsArrayHashTreeRoot computes the Merkle root of arrays of 32-byte hashes, such as [64][32]byte,
// according to the Simple Serialize specification used by QRL consensus.
func RootsArrayHashTreeRoot(vals [][]byte, length uint64) ([32]byte, error) {
	return ArraysRoot(vals, length)
}
