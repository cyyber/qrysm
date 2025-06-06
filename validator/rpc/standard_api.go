package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/common/hexutil"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	validatorServiceConfig "github.com/theQRL/qrysm/config/validator/service"
	"github.com/theQRL/qrysm/consensus-types/validator"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	zond "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	zondpbservice "github.com/theQRL/qrysm/proto/zond/service"
	"github.com/theQRL/qrysm/validator/client"
	"github.com/theQRL/qrysm/validator/keymanager"
	slashingprotection "github.com/theQRL/qrysm/validator/slashing-protection-history"
	"github.com/theQRL/qrysm/validator/slashing-protection-history/format"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListKeystores implements the standard validator key management API.
func (s *Server) ListKeystores(
	ctx context.Context, _ *empty.Empty,
) (*zondpbservice.ListKeystoresResponse, error) {
	if !s.walletInitialized {
		return nil, status.Error(codes.FailedPrecondition, "Qrysm Wallet not initialized. Please create a new wallet.")
	}
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready. Please try again once validator is ready.")
	}
	km, err := s.validatorService.Keymanager()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get Qrysm keymanager (possibly due to beacon node unavailable): %v", err)
	}
	if /*s.wallet.KeymanagerKind() != keymanager.Derived &&*/ s.wallet.KeymanagerKind() != keymanager.Local {
		return nil, status.Errorf(codes.FailedPrecondition, "Qrysm validator keys are not stored locally with this keymanager type.")
	}
	pubKeys, err := km.FetchValidatingPublicKeys(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve keystores: %v", err)
	}
	keystoreResponse := make([]*zondpbservice.ListKeystoresResponse_Keystore, len(pubKeys))
	for i := 0; i < len(pubKeys); i++ {
		keystoreResponse[i] = &zondpbservice.ListKeystoresResponse_Keystore{
			ValidatingPubkey: pubKeys[i][:],
		}
		/*
			if s.wallet.KeymanagerKind() == keymanager.Derived {
				keystoreResponse[i].DerivationPath = fmt.Sprintf(derived.ValidatingKeyDerivationPathTemplate, i)
			}
		*/
	}
	return &zondpbservice.ListKeystoresResponse{
		Data: keystoreResponse,
	}, nil
}

// ImportKeystores allows for importing keystores into Qrysm with their slashing protection history.
func (s *Server) ImportKeystores(
	ctx context.Context, req *zondpbservice.ImportKeystoresRequest,
) (*zondpbservice.ImportKeystoresResponse, error) {
	if !s.walletInitialized {
		statuses := groupImportErrors(req, "Qrysm Wallet not initialized. Please create a new wallet.")
		return &zondpbservice.ImportKeystoresResponse{Data: statuses}, nil
	}
	if s.validatorService == nil {
		statuses := groupImportErrors(req, "Validator service not ready. Please try again once validator is ready.")
		return &zondpbservice.ImportKeystoresResponse{Data: statuses}, nil
	}
	km, err := s.validatorService.Keymanager()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get keymanager (possibly due to beacon node unavailable): %v", err)
	}
	importer, ok := km.(keymanager.Importer)
	if !ok {
		statuses := groupImportErrors(req, "Keymanager kind cannot import keys")
		return &zondpbservice.ImportKeystoresResponse{Data: statuses}, nil
	}
	if len(req.Keystores) == 0 {
		return &zondpbservice.ImportKeystoresResponse{}, nil
	}
	keystores := make([]*keymanager.Keystore, len(req.Keystores))
	for i := 0; i < len(req.Keystores); i++ {
		k := &keymanager.Keystore{}
		err = json.Unmarshal([]byte(req.Keystores[i]), k)
		// if k.Description == "" && k.Name != "" {
		// 	k.Description = k.Name
		// }
		if err != nil {
			// we want to ignore unmarshal errors for now, proper status in importKeystore
			k.Pubkey = "invalid format"
		}
		keystores[i] = k
	}
	if req.SlashingProtection != "" {
		if err := slashingprotection.ImportStandardProtectionJSON(
			ctx, s.valDB, bytes.NewBuffer([]byte(req.SlashingProtection)),
		); err != nil {
			statuses := make([]*zondpbservice.ImportedKeystoreStatus, len(req.Keystores))
			for i := range statuses {
				statuses[i] = &zondpbservice.ImportedKeystoreStatus{
					Status:  zondpbservice.ImportedKeystoreStatus_ERROR,
					Message: fmt.Sprintf("could not import slashing protection: %v", err),
				}
			}
			return &zondpbservice.ImportKeystoresResponse{Data: statuses}, nil
		}
	}
	if len(req.Passwords) == 0 {
		req.Passwords = make([]string, len(req.Keystores))
	}

	// req.Passwords and req.Keystores are checked for 0 length in code above.
	if len(req.Passwords) > len(req.Keystores) {
		req.Passwords = req.Passwords[:len(req.Keystores)]
	}
	if len(req.Passwords) < len(req.Keystores) {
		passwordList := make([]string, len(req.Keystores))
		copy(passwordList, req.Passwords)
		req.Passwords = passwordList
	}

	statuses, err := importer.ImportKeystores(ctx, keystores, req.Passwords)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not import keystores: %v", err)
	}

	// If any of the keys imported had a slashing protection history before, we
	// stop marking them as deleted from our validator database.
	return &zondpbservice.ImportKeystoresResponse{Data: statuses}, nil
}

func groupImportErrors(req *zondpbservice.ImportKeystoresRequest, errorMessage string) []*zondpbservice.ImportedKeystoreStatus {
	statuses := make([]*zondpbservice.ImportedKeystoreStatus, len(req.Keystores))
	for i := 0; i < len(req.Keystores); i++ {
		statuses[i] = &zondpbservice.ImportedKeystoreStatus{
			Status:  zondpbservice.ImportedKeystoreStatus_ERROR,
			Message: errorMessage,
		}
	}
	return statuses
}

// DeleteKeystores allows for deleting specified public keys from Qrysm.
func (s *Server) DeleteKeystores(
	ctx context.Context, req *zondpbservice.DeleteKeystoresRequest,
) (*zondpbservice.DeleteKeystoresResponse, error) {
	if !s.walletInitialized {
		statuses := groupExportErrors(req, "Qrysm Wallet not initialized. Please create a new wallet.")
		return &zondpbservice.DeleteKeystoresResponse{Data: statuses}, nil
	}
	if s.validatorService == nil {
		statuses := groupExportErrors(req, "Validator service not ready")
		return &zondpbservice.DeleteKeystoresResponse{Data: statuses}, nil
	}
	km, err := s.validatorService.Keymanager()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get keymanager (possibly due to beacon node unavailable): %v", err)
	}
	if len(req.Pubkeys) == 0 {
		return &zondpbservice.DeleteKeystoresResponse{Data: make([]*zondpbservice.DeletedKeystoreStatus, 0)}, nil
	}
	statuses, err := km.DeleteKeystores(ctx, req.Pubkeys)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not delete keys: %v", err)
	}

	statuses, err = s.transformDeletedKeysStatuses(ctx, req.Pubkeys, statuses)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not transform deleted keys statuses: %v", err)
	}

	exportedHistory, err := s.slashingProtectionHistoryForDeletedKeys(ctx, req.Pubkeys, statuses)
	if err != nil {
		log.WithError(err).Warn("Could not get slashing protection history for deleted keys")
		statuses := groupExportErrors(req, "Non duplicate keys that were existing were deleted, but could not export slashing protection history.")
		return &zondpbservice.DeleteKeystoresResponse{Data: statuses}, nil
	}
	jsonHist, err := json.Marshal(exportedHistory)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Could not JSON marshal slashing protection history: %v",
			err,
		)
	}
	return &zondpbservice.DeleteKeystoresResponse{
		Data:               statuses,
		SlashingProtection: string(jsonHist),
	}, nil
}

func groupExportErrors(req *zondpbservice.DeleteKeystoresRequest, errorMessage string) []*zondpbservice.DeletedKeystoreStatus {
	statuses := make([]*zondpbservice.DeletedKeystoreStatus, len(req.Pubkeys))
	for i := 0; i < len(req.Pubkeys); i++ {
		statuses[i] = &zondpbservice.DeletedKeystoreStatus{
			Status:  zondpbservice.DeletedKeystoreStatus_ERROR,
			Message: errorMessage,
		}
	}
	return statuses
}

// For a list of deleted keystore statuses, we check if any NOT_FOUND status actually
// has a corresponding public key in the database. In this case, we transform the status
// to NOT_ACTIVE, as we do have slashing protection history for it and should not mark it
// as NOT_FOUND when returning a response to the caller.
func (s *Server) transformDeletedKeysStatuses(
	ctx context.Context, pubKeys [][]byte, statuses []*zondpbservice.DeletedKeystoreStatus,
) ([]*zondpbservice.DeletedKeystoreStatus, error) {
	pubKeysInDB, err := s.publicKeysInDB(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get public keys from DB: %v", err)
	}
	if len(pubKeysInDB) > 0 {
		for i := 0; i < len(pubKeys); i++ {
			keyExistsInDB := pubKeysInDB[bytesutil.ToBytes2592(pubKeys[i])]
			if keyExistsInDB && statuses[i].Status == zondpbservice.DeletedKeystoreStatus_NOT_FOUND {
				statuses[i].Status = zondpbservice.DeletedKeystoreStatus_NOT_ACTIVE
			}
		}
	}
	return statuses, nil
}

// Gets a map of all public keys in the database, useful for O(1) lookups.
func (s *Server) publicKeysInDB(ctx context.Context) (map[[field_params.DilithiumPubkeyLength]byte]bool, error) {
	pubKeysInDB := make(map[[field_params.DilithiumPubkeyLength]byte]bool)
	attestedPublicKeys, err := s.valDB.AttestedPublicKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get attested public keys from DB: %v", err)
	}
	proposedPublicKeys, err := s.valDB.ProposedPublicKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get proposed public keys from DB: %v", err)
	}
	for _, pk := range append(attestedPublicKeys, proposedPublicKeys...) {
		pubKeysInDB[pk] = true
	}
	return pubKeysInDB, nil
}

// Exports slashing protection data for a list of DELETED or NOT_ACTIVE keys only to be used
// as part of the DeleteKeystores endpoint.
func (s *Server) slashingProtectionHistoryForDeletedKeys(
	ctx context.Context, pubKeys [][]byte, statuses []*zondpbservice.DeletedKeystoreStatus,
) (*format.EIPSlashingProtectionFormat, error) {
	// We select the keys that were DELETED or NOT_ACTIVE from the previous action
	// and use that to filter our slashing protection export.
	filteredKeys := make([][]byte, 0, len(pubKeys))
	for i, pk := range pubKeys {
		if statuses[i].Status == zondpbservice.DeletedKeystoreStatus_DELETED ||
			statuses[i].Status == zondpbservice.DeletedKeystoreStatus_NOT_ACTIVE {
			filteredKeys = append(filteredKeys, pk)
		}
	}
	return slashingprotection.ExportStandardProtectionJSON(ctx, s.valDB, filteredKeys...)
}

/*
// ListRemoteKeys returns a list of all public keys defined for web3signer keymanager type.
func (s *Server) ListRemoteKeys(ctx context.Context, _ *empty.Empty) (*zondpbservice.ListRemoteKeysResponse, error) {
	if !s.walletInitialized {
		return nil, status.Error(codes.FailedPrecondition, "Qrysm Wallet not initialized. Please create a new wallet.")
	}
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready.")
	}
	km, err := s.validatorService.Keymanager()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get Qrysm keymanager (possibly due to beacon node unavailable): %v", err)
	}
	if s.wallet.KeymanagerKind() != keymanager.Web3Signer {
		return nil, status.Errorf(codes.FailedPrecondition, "Qrysm Wallet is not of type Web3Signer. Please execute validator client with web3signer flags.")
	}
	pubKeys, err := km.FetchValidatingPublicKeys(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve keystores: %v", err)
	}
	keystoreResponse := make([]*zondpbservice.ListRemoteKeysResponse_Keystore, len(pubKeys))
	for i := 0; i < len(pubKeys); i++ {
		keystoreResponse[i] = &zondpbservice.ListRemoteKeysResponse_Keystore{
			Pubkey:   pubKeys[i][:],
			Url:      s.validatorService.Web3SignerConfig.BaseEndpoint,
			Readonly: true,
		}
	}
	return &zondpbservice.ListRemoteKeysResponse{
		Data: keystoreResponse,
	}, nil
}

// ImportRemoteKeys imports a list of public keys defined for web3signer keymanager type.
func (s *Server) ImportRemoteKeys(ctx context.Context, req *zondpbservice.ImportRemoteKeysRequest) (*zondpbservice.ImportRemoteKeysResponse, error) {
	if !s.walletInitialized {
		return nil, status.Error(codes.FailedPrecondition, "Qrysm Wallet not initialized. Please create a new wallet.")
	}
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready.")
	}
	km, err := s.validatorService.Keymanager()
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Could not get Qrysm keymanager (possibly due to beacon node unavailable): %v", err))
	}
	if s.wallet.KeymanagerKind() != keymanager.Web3Signer {
		return nil, status.Errorf(codes.FailedPrecondition, "Qrysm Wallet is not of type Web3Signer. Please execute validator client with web3signer flags.")
	}
	adder, ok := km.(keymanager.PublicKeyAdder)
	if !ok {
		statuses := groupImportRemoteKeysErrors(req, "Keymanager kind cannot import public keys for web3signer keymanager type.")
		return &zondpbservice.ImportRemoteKeysResponse{Data: statuses}, nil
	}

	remoteKeys := make([][field_params.DilithiumPubkeyLength]byte, len(req.RemoteKeys))
	isUrlUsed := false
	for i, obj := range req.RemoteKeys {
		remoteKeys[i] = bytesutil.ToBytes2592(obj.Pubkey)
		if obj.Url != "" {
			isUrlUsed = true
		}
	}
	if isUrlUsed {
		log.Warnf("Setting web3signer base url for imported keys is not supported. Qrysm only uses the url from --validators-external-signer-url flag for web3signer.")
	}

	statuses, err := adder.AddPublicKeys(ctx, remoteKeys)
	if err != nil {
		sts := groupImportRemoteKeysErrors(req, fmt.Sprintf("Could not add keys;error: %v", err))
		return &zondpbservice.ImportRemoteKeysResponse{Data: sts}, nil
	}
	return &zondpbservice.ImportRemoteKeysResponse{
		Data: statuses,
	}, nil
}

func groupImportRemoteKeysErrors(req *zondpbservice.ImportRemoteKeysRequest, errorMessage string) []*zondpbservice.ImportedRemoteKeysStatus {
	statuses := make([]*zondpbservice.ImportedRemoteKeysStatus, len(req.RemoteKeys))
	for i := 0; i < len(req.RemoteKeys); i++ {
		statuses[i] = &zondpbservice.ImportedRemoteKeysStatus{
			Status:  zondpbservice.ImportedRemoteKeysStatus_ERROR,
			Message: errorMessage,
		}
	}
	return statuses
}

// DeleteRemoteKeys deletes a list of public keys defined for web3signer keymanager type.
func (s *Server) DeleteRemoteKeys(ctx context.Context, req *zondpbservice.DeleteRemoteKeysRequest) (*zondpbservice.DeleteRemoteKeysResponse, error) {
	if !s.walletInitialized {
		return nil, status.Error(codes.FailedPrecondition, "Qrysm Wallet not initialized. Please create a new wallet.")
	}
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready.")
	}
	km, err := s.validatorService.Keymanager()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get Qrysm keymanager (possibly due to beacon node unavailable): %v", err)
	}
	if s.wallet.KeymanagerKind() != keymanager.Web3Signer {
		return nil, status.Errorf(codes.FailedPrecondition, "Qrysm Wallet is not of type Web3Signer. Please execute validator client with web3signer flags.")
	}
	deleter, ok := km.(keymanager.PublicKeyDeleter)
	if !ok {
		statuses := groupDeleteRemoteKeysErrors(req, "Keymanager kind cannot delete public keys for web3signer keymanager type.")
		return &zondpbservice.DeleteRemoteKeysResponse{Data: statuses}, nil
	}
	remoteKeys := make([][field_params.DilithiumPubkeyLength]byte, len(req.Pubkeys))
	for i, key := range req.Pubkeys {
		remoteKeys[i] = bytesutil.ToBytes2592(key)
	}
	statuses, err := deleter.DeletePublicKeys(ctx, remoteKeys)
	if err != nil {
		sts := groupDeleteRemoteKeysErrors(req, fmt.Sprintf("Could not delete keys;error: %v", err))
		return &zondpbservice.DeleteRemoteKeysResponse{Data: sts}, nil
	}
	return &zondpbservice.DeleteRemoteKeysResponse{
		Data: statuses,
	}, nil
}
*/

/*
func groupDeleteRemoteKeysErrors(req *zondpbservice.DeleteRemoteKeysRequest, errorMessage string) []*zondpbservice.DeletedRemoteKeysStatus {
	statuses := make([]*zondpbservice.DeletedRemoteKeysStatus, len(req.Pubkeys))
	for i := 0; i < len(req.Pubkeys); i++ {
		statuses[i] = &zondpbservice.DeletedRemoteKeysStatus{
			Status:  zondpbservice.DeletedRemoteKeysStatus_ERROR,
			Message: errorMessage,
		}
	}
	return statuses
}
*/

func (s *Server) GetGasLimit(_ context.Context, req *zondpbservice.PubkeyRequest) (*zondpbservice.GetGasLimitResponse, error) {
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready")
	}
	validatorKey := req.Pubkey
	if err := validatePublicKey(validatorKey); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	resp := &zondpbservice.GetGasLimitResponse{
		Data: &zondpbservice.GetGasLimitResponse_GasLimit{
			Pubkey: validatorKey,
		},
	}
	settings := s.validatorService.ProposerSettings()
	if settings != nil {
		proposerOption, found := s.validatorService.ProposerSettings().ProposeConfig[bytesutil.ToBytes2592(validatorKey)]
		if found {
			if proposerOption.BuilderConfig != nil {
				resp.Data.GasLimit = uint64(proposerOption.BuilderConfig.GasLimit)
				return resp, nil
			}
		} else if s.validatorService.ProposerSettings().DefaultConfig != nil && s.validatorService.ProposerSettings().DefaultConfig.BuilderConfig != nil {
			resp.Data.GasLimit = uint64(s.validatorService.ProposerSettings().DefaultConfig.BuilderConfig.GasLimit)
			return resp, nil
		}
	}
	resp.Data.GasLimit = params.BeaconConfig().DefaultBuilderGasLimit
	return resp, nil
}

// SetGasLimit updates GasLimt of the public key.
func (s *Server) SetGasLimit(ctx context.Context, req *zondpbservice.SetGasLimitRequest) (*empty.Empty, error) {
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready")
	}
	validatorKey := req.Pubkey

	if err := validatePublicKey(validatorKey); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	settings := s.validatorService.ProposerSettings()
	if settings == nil {
		return &empty.Empty{}, status.Errorf(codes.FailedPrecondition, "no proposer settings were found to update")
	} else if settings.ProposeConfig == nil {
		if settings.DefaultConfig == nil || settings.DefaultConfig.BuilderConfig == nil || !settings.DefaultConfig.BuilderConfig.Enabled {
			return &empty.Empty{}, status.Errorf(codes.FailedPrecondition, "gas limit changes only apply when builder is enabled")
		}
		settings.ProposeConfig = make(map[[field_params.DilithiumPubkeyLength]byte]*validatorServiceConfig.ProposerOption)
		option := settings.DefaultConfig.Clone()
		option.BuilderConfig.GasLimit = validator.Uint64(req.GasLimit)
		settings.ProposeConfig[bytesutil.ToBytes2592(validatorKey)] = option
	} else {
		proposerOption, found := settings.ProposeConfig[bytesutil.ToBytes2592(validatorKey)]
		if found {
			if proposerOption.BuilderConfig == nil || !proposerOption.BuilderConfig.Enabled {
				return &empty.Empty{}, status.Errorf(codes.FailedPrecondition, "gas limit changes only apply when builder is enabled")
			} else {
				proposerOption.BuilderConfig.GasLimit = validator.Uint64(req.GasLimit)
			}
		} else {
			if settings.DefaultConfig == nil {
				return &empty.Empty{}, status.Errorf(codes.FailedPrecondition, "gas limit changes only apply when builder is enabled")
			}
			option := settings.DefaultConfig.Clone()
			option.BuilderConfig.GasLimit = validator.Uint64(req.GasLimit)
			settings.ProposeConfig[bytesutil.ToBytes2592(validatorKey)] = option
		}
	}
	// save the settings
	if err := s.validatorService.SetProposerSettings(ctx, settings); err != nil {
		return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set proposer settings: %v", err)
	}
	// override the 200 success with 202 according to the specs
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "202")); err != nil {
		return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set custom success code header: %v", err)
	}

	return &empty.Empty{}, nil
}

func (s *Server) DeleteGasLimit(ctx context.Context, req *zondpbservice.DeleteGasLimitRequest) (*empty.Empty, error) {
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready")
	}
	validatorKey := req.Pubkey
	if err := validatePublicKey(validatorKey); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	proposerSettings := s.validatorService.ProposerSettings()
	if proposerSettings != nil && proposerSettings.ProposeConfig != nil {
		proposerOption, found := proposerSettings.ProposeConfig[bytesutil.ToBytes2592(validatorKey)]
		if found && proposerOption.BuilderConfig != nil {
			// If proposerSettings has default value, use it.
			if proposerSettings.DefaultConfig != nil && proposerSettings.DefaultConfig.BuilderConfig != nil {
				proposerOption.BuilderConfig.GasLimit = proposerSettings.DefaultConfig.BuilderConfig.GasLimit
			} else {
				// Fallback to using global default.
				proposerOption.BuilderConfig.GasLimit = validator.Uint64(params.BeaconConfig().DefaultBuilderGasLimit)
			}
			// save the settings
			if err := s.validatorService.SetProposerSettings(ctx, proposerSettings); err != nil {
				return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set proposer settings: %v", err)
			}
			// Successfully deleted gas limit (reset to proposer config default or global default).
			// Return with success http code "204".
			if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "204")); err != nil {
				return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set custom http code 204 header: %v", err)
			}
			return &empty.Empty{}, nil
		}
	}
	// Otherwise, either no proposerOption is found for the pubkey or proposerOption.BuilderConfig is not enabled at all,
	// we response "not found".
	return nil, status.Error(codes.NotFound, fmt.Sprintf("no gaslimt found for pubkey: %q", hexutil.Encode(validatorKey)))
}

// ListFeeRecipientByPubkey returns the public key to zond address mapping object to the end user.
func (s *Server) ListFeeRecipientByPubkey(ctx context.Context, req *zondpbservice.PubkeyRequest) (*zondpbservice.GetFeeRecipientByPubkeyResponse, error) {
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready")
	}

	validatorKey := req.Pubkey
	if err := validatePublicKey(validatorKey); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	finalResp := &zondpbservice.GetFeeRecipientByPubkeyResponse{
		Data: &zondpbservice.GetFeeRecipientByPubkeyResponse_FeeRecipient{
			Pubkey: validatorKey,
		},
	}

	proposerSettings := s.validatorService.ProposerSettings()

	// If fee recipient is defined for this specific pubkey in proposer configuration, use it
	if proposerSettings != nil && proposerSettings.ProposeConfig != nil {
		proposerOption, found := proposerSettings.ProposeConfig[bytesutil.ToBytes2592(validatorKey)]

		if found && proposerOption.FeeRecipientConfig != nil {
			finalResp.Data.Zondaddress = proposerOption.FeeRecipientConfig.FeeRecipient.Bytes()
			return finalResp, nil
		}
	}

	// If fee recipient is defined in default configuration, use it
	if proposerSettings != nil && proposerSettings.DefaultConfig != nil && proposerSettings.DefaultConfig.FeeRecipientConfig != nil {
		finalResp.Data.Zondaddress = proposerSettings.DefaultConfig.FeeRecipientConfig.FeeRecipient.Bytes()
		return finalResp, nil
	}

	// Else, use the one defined in beacon node TODO: remove this with db removal
	resp, err := s.beaconNodeValidatorClient.GetFeeRecipientByPubKey(ctx, &zond.FeeRecipientByPubKeyRequest{
		PublicKey: validatorKey,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to retrieve default fee recipient from beacon node")
	}

	if resp != nil && len(resp.FeeRecipient) != 0 {
		finalResp.Data.Zondaddress = resp.FeeRecipient
		return finalResp, nil
	}

	return nil, status.Error(codes.InvalidArgument, "No fee recipient set")
}

// SetFeeRecipientByPubkey updates the zond address mapped to the public key.
func (s *Server) SetFeeRecipientByPubkey(ctx context.Context, req *zondpbservice.SetFeeRecipientByPubkeyRequest) (*empty.Empty, error) {
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready")
	}

	validatorKey := req.Pubkey
	feeRecipient := common.BytesToAddress(req.Zondaddress)

	if err := validatePublicKey(validatorKey); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	encoded := hexutil.EncodeZ(req.Zondaddress)

	if !common.IsAddress(encoded) {
		return nil, status.Error(
			codes.InvalidArgument, "Fee recipient is not a valid Zond address")
	}
	settings := s.validatorService.ProposerSettings()
	switch {
	case settings == nil:
		settings = &validatorServiceConfig.ProposerSettings{
			ProposeConfig: map[[field_params.DilithiumPubkeyLength]byte]*validatorServiceConfig.ProposerOption{
				bytesutil.ToBytes2592(validatorKey): {
					FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
						FeeRecipient: feeRecipient,
					},
					BuilderConfig: nil,
				},
			},
			DefaultConfig: nil,
		}
	case settings.ProposeConfig == nil:
		var builderConfig *validatorServiceConfig.BuilderConfig
		if settings.DefaultConfig != nil && settings.DefaultConfig.BuilderConfig != nil {
			builderConfig = settings.DefaultConfig.BuilderConfig.Clone()
		}
		settings.ProposeConfig = map[[field_params.DilithiumPubkeyLength]byte]*validatorServiceConfig.ProposerOption{
			bytesutil.ToBytes2592(validatorKey): {
				FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
					FeeRecipient: feeRecipient,
				},
				BuilderConfig: builderConfig,
			},
		}
	default:
		proposerOption, found := settings.ProposeConfig[bytesutil.ToBytes2592(validatorKey)]
		if found && proposerOption != nil {
			proposerOption.FeeRecipientConfig = &validatorServiceConfig.FeeRecipientConfig{
				FeeRecipient: feeRecipient,
			}
		} else {
			var builderConfig = &validatorServiceConfig.BuilderConfig{}
			if settings.DefaultConfig != nil && settings.DefaultConfig.BuilderConfig != nil {
				builderConfig = settings.DefaultConfig.BuilderConfig.Clone()
			}
			settings.ProposeConfig[bytesutil.ToBytes2592(validatorKey)] = &validatorServiceConfig.ProposerOption{
				FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
					FeeRecipient: feeRecipient,
				},
				BuilderConfig: builderConfig,
			}
		}
	}
	// save the settings
	if err := s.validatorService.SetProposerSettings(ctx, settings); err != nil {
		return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set proposer settings: %v", err)
	}
	// override the 200 success with 202 according to the specs
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "202")); err != nil {
		return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set custom success code header: %v", err)
	}
	return &empty.Empty{}, nil
}

// DeleteFeeRecipientByPubkey updates the zond address mapped to the public key to the default fee recipient listed
func (s *Server) DeleteFeeRecipientByPubkey(ctx context.Context, req *zondpbservice.PubkeyRequest) (*empty.Empty, error) {
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready")
	}

	validatorKey := req.Pubkey

	if err := validatePublicKey(validatorKey); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	settings := s.validatorService.ProposerSettings()

	if settings != nil && settings.ProposeConfig != nil {
		proposerOption, found := settings.ProposeConfig[bytesutil.ToBytes2592(validatorKey)]
		if found {
			proposerOption.FeeRecipientConfig = nil
		}
	}

	// save the settings
	if err := s.validatorService.SetProposerSettings(ctx, settings); err != nil {
		return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set proposer settings: %v", err)
	}

	// override the 200 success with 204 according to the specs
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "204")); err != nil {
		return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set custom success code header: %v", err)
	}
	return &empty.Empty{}, nil
}

func validatePublicKey(pubkey []byte) error {
	if len(pubkey) != field_params.DilithiumPubkeyLength {
		return status.Errorf(
			codes.InvalidArgument, "Provided public key in path is not byte length %d and not a valid dilithium public key", field_params.DilithiumPubkeyLength)
	}
	return nil
}

// SetVoluntaryExit creates a signed voluntary exit message and returns a VoluntaryExit object.
func (s *Server) SetVoluntaryExit(ctx context.Context, req *zondpbservice.SetVoluntaryExitRequest) (*zondpbservice.SetVoluntaryExitResponse, error) {
	if s.validatorService == nil {
		return nil, status.Error(codes.FailedPrecondition, "Validator service not ready")
	}
	if err := validatePublicKey(req.Pubkey); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if s.wallet == nil {
		return nil, status.Error(codes.FailedPrecondition, "No wallet found")
	}
	km, err := s.validatorService.Keymanager()
	if err != nil {
		return nil, err
	}
	if req.Epoch == 0 {
		genesisResponse, err := s.beaconNodeClient.GetGenesis(ctx, &emptypb.Empty{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not create voluntary exit: %v", err)
		}
		epoch, err := client.CurrentEpoch(genesisResponse.GenesisTime)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "gRPC call to get genesis time failed: %v", err)
		}
		req.Epoch = epoch
	}
	sve, err := client.CreateSignedVoluntaryExit(
		ctx,
		s.beaconNodeValidatorClient,
		km.Sign,
		req.Pubkey,
		req.Epoch,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not create voluntary exit: %v", err)
	}

	return &zondpbservice.SetVoluntaryExitResponse{
		Data: &zondpbservice.SetVoluntaryExitResponse_SignedVoluntaryExit{
			Message: &zondpbservice.SetVoluntaryExitResponse_SignedVoluntaryExit_VoluntaryExit{
				Epoch:          uint64(sve.Exit.Epoch),
				ValidatorIndex: uint64(sve.Exit.ValidatorIndex),
			},
			Signature: sve.Signature,
		},
	}, nil
}
