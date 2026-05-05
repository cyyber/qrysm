//go:build develop

package blockchain

// stopTests is consumed by TestMain in blockchain_test.go. The `!develop`
// variant flips it to true so default `go test ./...` runs short-circuit.
var stopTests = false
