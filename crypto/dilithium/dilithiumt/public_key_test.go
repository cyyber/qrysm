package dilithiumt_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/qrysm/v4/crypto/dilithium/dilithiumt"
	"github.com/theQRL/qrysm/v4/testing/assert"
	"github.com/theQRL/qrysm/v4/testing/require"
)

func TestPublicKeyFromBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		err   error
	}{
		{
			name: "Nil",
			err:  errors.New("public key must be 2592 bytes"),
		},
		{
			name:  "Empty",
			input: []byte{},
			err:   errors.New("public key must be 2592 bytes"),
		},
		{
			name:  "Short",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			err:   errors.New("public key must be 2592 bytes"),
		},
		{
			name:  "Long",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			err:   errors.New("public key must be 2592 bytes"),
		},
		{
			name:  "Good",
			input: ezDecode(t, "0xe0a586bb51db522c31abcbce14e6cbf6a5bbc7b3331cdb76378ca1b98acff048c11099c2713f229c349c430a6aa5623fab8d39ec266e0e7d81543fc2e4b905ec7fba75b9ab3aae53e18e2a018297ebe4bb2d0a22bd13b60b938461d922ec81dfe152224c51abcccd4105799ee2b70b53cf2401a3c01664c20ab368c4c3ccc764be5063488750f79480adcac444e274fb46500aeb2449d2a81e44c3528c70554a9ecd5b25b39550d469a43f5ec2afce668aa6598aa1c5618e569bdf08ec700a21950d6d2df3337ff196b6fcb53de94e7e127dbd7edf9c5df70c41c715b48cf4e5ab5d0e1bc4d9ead578150f98244ea47dba29af25b12a72054618d0341ebbae8e5c61cf6583c0151fdcf1323ee3cd65f8f739dc621f2aa77f8dfe36a7cef15162972c25a193bd306918deb8d6395367586ee6a534340c07caf6496dc393a0189cf81325499132a012a2b8a6152be3d3d010aadba896af83d9d447741a66100f72da46c9282a8a9af5bfec0d84d88882ce0a090147dbcef2f100f8744094a8e3712c26d875996f56a92fd99a39537197f0bbe58bb706061426e62406a300626f64b7dd813c756c159cea82a6cf82b9890be40284720b9aa9c6f1a3a78bfb8607b438ec3665225fe21370770cfdbbc20ef6525362d413ab23e85d5f6ef38a43b44874828137ab977dd9a145913ecffe2700a225042b766158d26288434511014928efdb857df4217430e18bd6c370c8327b4451611c66a118193f1155ffeef32d9b26b02d04cd083964f53b59b5ffd02789be6e8aeae4f615afb39e53f5cfbf3d9ec23640fea711fc6751abe9b3606959ac12aeb827a76a515f27d0e0f1e003e00a91a1d20b97ecde53202d6f9d61a1d4b0bd7d4f0622c2a90d67ba40f59a450191aa340bcc7b3b3107830ce1fb791a97930fd68c6b9ea0848c0591edb0a6302e0984d7f096ccb980803bbeaae3550d8996a001ecba956d3c2bb20eaa33094071e639983693f64809e449c29b59bc0b4f1530ec273f366db337bc64d95a9e26cc21ed0685cb2c606b994505bdd6237dfdb414df7eee7544f34cdf5f0e6ca1e5280b493446cb883413e26e06a00354bced7a5fd410fd92ffc39443d9e8f208aec8d81d958c060203bbe75db0cb2b982524b5e91135d4ef671ebe6c55c24bdb00d89b78c7d8fed674d1fac6d6d61bb671a996d3efe27a254e40967cb60c3c7ac5814ca5e5768f268c7002ba200da9200fc5498d07833c4b25a111d35f64cd26b108a897616d4324984e0833937344b904b964d5f292eeba6075987b5cc092bd40697ef9b2ea95cbc1eed5bf3f6337145351e98291853c3bf75eb1a533817cab5dc8d87abf034696e8dc9ad20089d79086b8608cb07101b62ae744beba3cc71e75701d46c35f317ecd1f3bb6d6078bd8cf25a55bbd200d7bfb5c9e3e2167a6abe8a6636ca82bdd63c3007b39d9a57b9f258ee4bb94325b20744087ba3a2bbda513ed067b003a0d6a4197bae5776cb25899911f92e3cdd779931e98cb11846dac49480af3c2fc3596825ccbb7a7dfa3e714d8fc809acc57577dd448e477bff03a907f8410f2ae12a9ab4a3d7738315c07f42e5af416560aafa035ae4d4b72d5e59a45dcd4c91000cd8cef454c7a157276cb2610d2d08a7bc90550c85e317fdb2d83ee26f49198dd035bbe39d6eedfbc91a1cffb5682f0410204c281f3cb8d702258c77214d77e92f1e5f4db2a6be18911c5f3950a3228d1850722f4ce0a5d56a8acf2e0311290e1334a2bfbf1251d6cb46f2a028aaa7be144f38fdb222e8a2d6320f98796731847b2449774c025a452c72dfbb9f05959c88ed86256f5fcea5458c3e22340d8ad3c3ff548f03346c55f74d6ab3aff1311a302bb8c5cd55b44528fd08ecab030c1e47385ed27de5819fd798ada8858462de7fd55aa7239e03079d976669e52ff22bf9ab6fa4860064dd5033ca6ae1fec5c628e5bc2b190ae5483514d841a25b04d127d9c536e32f3bada7b46cbb14b5718c88ed826a8c19d1fd43a7f7ff6860a88adf9fcf1415eb2c56e12a7dc6a73e24bd7cbcc7fa39ce7358f20736adc11842e72a5107bcbe78f56bb56fe403d51ea531d7d4f2681fd05f5326d7e5dd7e889a3380f9dfe8124d8f258d6f9ba6f0f6467e787d996da6310196a70f551e64d1a9dc51fe907227f43a1fb54a572db183edb726375a7a1096daeb8d7b069bf8886d282dabdb7e9101fadcdfb23c57be75a193cc3459401d14836d250b197e6ae0e4818b2bea75db388bf36f311eff18ac14b9f0fe1a354d8d397439fd202d61d545f430676eb16c6ecc4c3f583fa8767d65cdc4f3155af47629cf1b0b833a12391b02b1781f1c31cb6b05160241b1e02c5889db631ad2fa905d608b2831b45529dd7550d5ff91d4b7ae23533b1f6875b38be0f26f4479cb75579d8612ff2cfec981344598c76054584f6350d296c2436e2d43f184556d4e6208483e010ba8bcdf413d659fab3353bb8f53f085dfe24910f28b82ae382047383e81922f2d05b13d073b3fbb8042c9c1dd6a073109afffac32117a6b4162387949a9c2b21661eb321a340978b4c43dcb8ce264d6e30751c1e91551f4c2efec349bf0f083db63f3bbbbc83be7eb044b17a7fbfdeedbdd80e76580a5082d7534cea34620997ee593fc0c725a9cc41f192cdcb85d2021f2dafea48f14f63d01329845c0533210075ac3d1b674a5535d37c5c5acbd8fdee0ef9d3dc66b9fdec661f3ed53d1c70c825937716af2d44510d07876b3d52c063e7ddca41faee15d3b81940dd50d41ad5791b4b37f44cecc11db9ce58c7555491ee822e8ff1d8b0dac3eb409f8b827561ea6b7d88af82892a53ff2d239c76a8a1a717b101c7b9d7db85a84a508276b8d1ba972d31089cce5dba3ed722ead0849d336b1002f41f1b1b93d2a7e56e5c222d21327d872534aae80e8f7020c4fda6fd4765bd94df4aa38c51924b356412ae0ceff6adfdad9b9c793ee6aec73a902f658ffd6af25abf374368e38a8b9e91b34d2eeca566eb39ffa67978077870b21279afc7760f38639ad6fa152af670f25de919fcecc16755bafff466a0b8d9910bf84bc5917a33ed76fd62c47a9a2ca68055668a13f11616b7f95cda26c2b09bbe8c83609af99ec41470bda5b12524a849950caf6fb96d908dabca97187858c83a54cb2dee7fcabbc0fea8d3ca1b860d1d7b5eb1dc2a687330d2fb237f55d97fbab4694e4037355548f1c20122da77eb0b7b90205989bb9ef52f76f88770eabf56f5d9ccaf572b3eadb7c9810e93b675e7e9ea26f8d8749fcb23c63993d62406db2a53996dc053e698e70360492e2c467e1baf2d76a9dc74f23c3be3d27685c5fd07a30d2aab3f2cd7fde563e29a3434ee6a51f795a5e114a3d6259732362126da789d82ee54dae91c3e2c060a4f79943068cb6a3ee5692587a67816aa5a9c5ff3805173c72a5ad2b0ebd8588253bae50da117d938901f8ffbf725ced16a76f9f53d782ebe1f0d6f6dfdaa4fe8f93ec6246b66e561c740fc7eaf6c771659e90f545b9e89221fc9450543424f0a14ad7484253251f658e56cf1cb161b4cee63c6c5b96cf8c06e6aa524c8209205de7fdbf1e233135755ed6300ed4c096764fe4dd4855f421d272cd63150db47bc6f47bf624798"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := dilithiumt.PublicKeyFromBytes(test.input)
			if test.err != nil {
				assert.NotEqual(t, nil, err, "No error returned")
				assert.ErrorContains(t, test.err.Error(), err, "Unexpected error returned")
			} else {
				assert.NoError(t, err)
				assert.DeepEqual(t, 0, bytes.Compare(res.Marshal(), test.input))
			}
		})
	}
}

func TestPublicKey_Copy(t *testing.T) {
	priv, err := dilithiumt.RandKey()
	require.NoError(t, err)
	pubkeyA := priv.PublicKey()
	pubkeyBytes := pubkeyA.Marshal()

	pubkeyB := pubkeyA.Copy()
	// TODO(rgeraldes24): remove
	// pubkeyB := pubkeyA.Copy()
	// priv2, err := dilithiumt.RandKey()
	// require.NoError(t, err)
	// pubkeyB.Aggregate(priv2.PublicKey())

	require.DeepEqual(t, pubkeyA.Marshal(), pubkeyBytes, "Pubkey was mutated after copy")
	require.DeepEqual(t, pubkeyA, pubkeyB)
}

func BenchmarkPublicKeyFromBytes(b *testing.B) {
	priv, err := dilithiumt.RandKey()
	require.NoError(b, err)
	pubkey := priv.PublicKey()
	pubkeyBytes := pubkey.Marshal()

	b.Run("cache on", func(b *testing.B) {
		dilithiumt.EnableCaches()
		for i := 0; i < b.N; i++ {
			_, err := dilithiumt.PublicKeyFromBytes(pubkeyBytes)
			require.NoError(b, err)
		}
	})

	b.Run("cache off", func(b *testing.B) {
		dilithiumt.DisableCaches()
		for i := 0; i < b.N; i++ {
			_, err := dilithiumt.PublicKeyFromBytes(pubkeyBytes)
			require.NoError(b, err)
		}
	})

}

func ezDecode(t *testing.T, s string) []byte {
	v, err := hexutil.Decode(s)
	require.NoError(t, err)
	return v
}