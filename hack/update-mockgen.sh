#!/bin/bash

# Script to update mock files after proto/qrysm/v1alpha1/services.proto changes.
# Use a space to separate mock destination from its interfaces.

mock_path="testing/mock"
iface_mock_path="testing/validator-mock"

# github.com/theQRL/qrysm/proto/qrysm/v1alpha1
# ------------------------------------------------------
proto_mocks_v1alpha1=(
      "$mock_path/beacon_service_mock.go BeaconChainClient"
      "$mock_path/beacon_validator_server_mock.go BeaconNodeValidatorServer,BeaconNodeValidator_WaitForActivationServer,BeaconNodeValidator_WaitForChainStartServer"
      "$mock_path/beacon_validator_client_mock.go BeaconNodeValidatorClient,BeaconNodeValidator_WaitForChainStartClient,BeaconNodeValidator_WaitForActivationClient"
      "$mock_path/node_service_mock.go NodeClient"
)

for ((i = 0; i < ${#proto_mocks_v1alpha1[@]}; i++)); do
    file=${proto_mocks_v1alpha1[i]% *};
    interfaces=${proto_mocks_v1alpha1[i]#* };
    echo "generating $file for interfaces: $interfaces";
    echo
    GO11MODULE=on mockgen -package=mock -destination="$file" github.com/theQRL/qrysm/proto/qrysm/v1alpha1 "$interfaces"
done

# github.com/theQRL/qrysm/proto/zond/service
# ---------------------------------------------------
proto_mocks_service=(
      "$mock_path/event_service_mock.go EventsClient,Events_StreamEventsClient,Events_StreamEventsServer"
)

for ((i = 0; i < ${#proto_mocks_service[@]}; i++)); do
    file=${proto_mocks_service[i]% *};
    interfaces=${proto_mocks_service[i]#* };
    echo "generating $file for interfaces: $interfaces";
    echo
    GO11MODULE=on mockgen -package=mock -destination="$file" github.com/theQRL/qrysm/proto/zond/service "$interfaces"
done

# github.com/theQRL/qrysm/proto/qrysm/v1alpha1/validator-client
# -----------------------------------------------------------------------
proto_mocks_v1alpha1_validator_clients=(
    "$mock_path/keymanager_mock.go RemoteSignerClient"
)

for ((i = 0; i < ${#proto_mocks_v1alpha1_validator_clients[@]}; i++)); do
    file=${proto_mocks_v1alpha1_validator_clients[i]% *};
    interfaces=${proto_mocks_v1alpha1_validator_clients[i]#* };
    echo "generating $file for interfaces: $interfaces";
    echo
    GO11MODULE=on mockgen -package=mock -destination="$file" github.com/theQRL/qrysm/proto/qrysm/v1alpha1/validator-client "$interfaces"
done

# github.com/theQRL/qrysm/validator/client/iface
# --------------------------------------------------------
iface_mocks=(
      "$iface_mock_path/beacon_chain_client_mock.go BeaconChainClient"
      "$iface_mock_path/node_client_mock.go NodeClient"
      "$iface_mock_path/validator_client_mock.go ValidatorClient"
)

for ((i = 0; i < ${#iface_mocks[@]}; i++)); do
    file=${iface_mocks[i]% *};
    interfaces=${iface_mocks[i]#* };
    echo "generating $file for interfaces: $interfaces";
    GO11MODULE=on mockgen -package=validator_mock -destination="$file" github.com/theQRL/qrysm/validator/client/iface "$interfaces"
done

goimports -w "$mock_path/."
gofmt -s -w "$mock_path/."

# github.com/theQRL/qrysm/validator/client/beacon-api
# -------------------------------------------------------------
beacon_api_mock_path="validator/client/beacon-api/mock"
beacon_api_mocks=(
      "$beacon_api_mock_path/genesis_mock.go genesis.go"
      "$beacon_api_mock_path/duties_mock.go duties.go"
      "$beacon_api_mock_path/json_rest_handler_mock.go json_rest_handler.go"
      "$beacon_api_mock_path/state_validators_mock.go state_validators.go"
      "$beacon_api_mock_path/beacon_block_converter_mock.go beacon_block_converter.go"
)

for ((i = 0; i < ${#beacon_api_mocks[@]}; i++)); do
    file=${beacon_api_mocks[i]% *};
    source=${beacon_api_mocks[i]#* };
    echo "generating $file for file: $source";
    GO11MODULE=on mockgen -package=mock -source="validator/client/beacon-api/$source" -destination="$file"
done

goimports -w "$beacon_api_mock_path/."
gofmt -s -w "$beacon_api_mock_path/."

# github.com/theQRL/qrysm/crypto/dilithium
# --------------------------------------------
crypto_dilithium_common_mock_path="crypto/dilithium/common/mock"
crypto_dilithium_common_mocks=(
      "$crypto_dilithium_common_mock_path/interface_mock.go interface.go"
)

for ((i = 0; i < ${#crypto_dilithium_common_mocks[@]}; i++)); do
    file=${crypto_dilithium_common_mocks[i]% *};
    source=${crypto_dilithium_common_mocks[i]#* };
    echo "generating $file for file: $source";
    GO11MODULE=on mockgen -package=mock -source="crypto/dilithium/common/$source" -destination="$file"
done

# goimports -w "$crypto_dilithium_common_mock_path/."
# gofmt -s -w "$crypto_dilithium_common_mock_path/."
