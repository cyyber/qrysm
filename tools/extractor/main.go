package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/theQRL/qrysm/beacon-chain/core/transition/interop"
	"github.com/theQRL/qrysm/beacon-chain/db"
	"github.com/theQRL/qrysm/config/features"
	"github.com/theQRL/qrysm/consensus-types/primitives"
)

var (
	// Required fields
	datadir = flag.String("datadir", "", "Path to data directory.")

	state = flag.Uint("state", 0, "Extract state at this slot.")
)

func main() {
	resetCfg := features.InitWithReset(&features.Flags{WriteSSZStateTransitions: true})
	defer resetCfg()
	flag.Parse()
	fmt.Println("Starting process...")
	d, err := db.NewDB(context.Background(), *datadir)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	slot := primitives.Slot(*state)
	_, roots, err := d.BlockRootsBySlot(ctx, slot)
	if err != nil {
		panic(err)
	}
	if len(roots) != 1 {
		fmt.Printf("Expected 1 block root for slot %d, got %d roots", *state, len(roots))
	}
	s, err := d.State(ctx, roots[0])
	if err != nil {
		panic(err)
	}

	interop.WriteStateToDisk(s)
	fmt.Println("done")
}
