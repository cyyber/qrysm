package p2p

import (
	"reflect"
	"testing"

	"github.com/theQRL/qrysm/v4/config/params"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/testing/assert"
)

func TestMappingHasNoDuplicates(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	m := make(map[reflect.Type]bool)
	for _, v := range gossipTopicMappings {
		if _, ok := m[reflect.TypeOf(v)]; ok {
			t.Errorf("%T is duplicated in the topic mapping", v)
		}
		m[reflect.TypeOf(v)] = true
	}
}

func TestGossipTopicMappings_CorrectBlockType(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	bCfg := params.BeaconConfig().Copy()
	//altairForkEpoch := primitives.Epoch(100)
	//BellatrixForkEpoch := primitives.Epoch(200)
	//CapellaForkEpoch := primitives.Epoch(300)

	//bCfg.AltairForkEpoch = altairForkEpoch
	//bCfg.BellatrixForkEpoch = BellatrixForkEpoch
	//bCfg.CapellaForkEpoch = CapellaForkEpoch
	//bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.AltairForkVersion)] = primitives.Epoch(100)
	//bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.BellatrixForkVersion)] = primitives.Epoch(200)
	//bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.CapellaForkVersion)] = primitives.Epoch(300)
	params.OverrideBeaconConfig(bCfg)

	// Capella Fork
	pMessage := GossipTopicMappings(BlockSubnetTopicFormat, 0)
	_, ok := pMessage.(*zondpb.SignedBeaconBlock)
	assert.Equal(t, true, ok)

	/*
		// Altair Fork
		pMessage = GossipTopicMappings(BlockSubnetTopicFormat, altairForkEpoch)
		_, ok = pMessage.(*zondpb.SignedBeaconBlockAltair)
		assert.Equal(t, true, ok)

		// Bellatrix Fork
		pMessage = GossipTopicMappings(BlockSubnetTopicFormat, BellatrixForkEpoch)
		_, ok = pMessage.(*zondpb.SignedBeaconBlockBellatrix)
		assert.Equal(t, true, ok)

		// Capella Fork
		pMessage = GossipTopicMappings(BlockSubnetTopicFormat, CapellaForkEpoch)
		_, ok = pMessage.(*zondpb.SignedBeaconBlockCapella)
		assert.Equal(t, true, ok)
	*/
}
