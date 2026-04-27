//go:build !develop

package blockchain

// stopTests, when true, makes blockchain_test.go's TestMain exit before
// running anything. The `develop` tag swaps in a no-op alternative (see
// checktags_develop_test.go) so the real tests run.
var stopTests = true
