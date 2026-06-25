# End-to-end Testing Package

This is the main project folder of the end-to-end testing suite for Qrysm. This performs a full end-to-end test for Qrysm, including spinning up execution clients, sending deposits to the deposit contract, and making sure the beacon node and its validators are running and performing properly for a few epochs.
It also performs a test on a syncing node, and supports feature flags to allow easy E2E testing of experimental features. 

## How it works

Run the Bazel E2E targets in this package for scenario-specific Qrysm coverage.
