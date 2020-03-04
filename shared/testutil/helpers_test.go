package testutil

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/shared/params"
)

func TestBlockSignature(t *testing.T) {
	beaconState, privKeys := DeterministicGenesisState(t, 100)
	block, err := GenerateFullBlock(beaconState, privKeys, nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	beaconState.SetSlot(beaconState.Slot() + 1)
	proposerIdx, err := helpers.BeaconProposerIndex(beaconState)
	if err != nil {
		t.Error(err)
	}
	beaconState.SetSlot(beaconState.Slot() - 1)
	epoch := helpers.SlotToEpoch(block.Block.Slot)
	domain, err := helpers.Domain(beaconState.Fork(), epoch, params.BeaconConfig().DomainBeaconProposer)
	if err != nil {
		t.Fatal(err)
	}
	signingRoot, err := helpers.ComputeSigningRoot(block.Block, domain)
	if err != nil {
		t.Error(err)
	}

	blockSig := privKeys[proposerIdx].Sign(signingRoot[:]).Marshal()

	signature, err := BlockSignature(beaconState, block.Block, privKeys)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(blockSig[:], signature.Marshal()) {
		t.Errorf("Expected block signatures to be equal, received %#x != %#x", blockSig[:], signature.Marshal())
	}
}

func TestRandaoReveal(t *testing.T) {
	beaconState, privKeys := DeterministicGenesisState(t, 100)

	epoch := helpers.CurrentEpoch(beaconState)
	randaoReveal, err := RandaoReveal(beaconState, epoch, privKeys)
	if err != nil {
		t.Error(err)
	}

	proposerIdx, err := helpers.BeaconProposerIndex(beaconState)
	if err != nil {
		t.Error(err)
	}
	buf := make([]byte, 32)
	binary.LittleEndian.PutUint64(buf, epoch)
	domain, err := helpers.Domain(beaconState.Fork(), epoch, params.BeaconConfig().DomainRandao)
	if err != nil {
		t.Fatal(err)
	}
	root, err := helpers.ComputeSigningRoot(epoch, domain)
	if err != nil {
		t.Fatal(err)
	}
	// We make the previous validator's index sign the message instead of the proposer.
	epochSignature := privKeys[proposerIdx].Sign(root[:]).Marshal()

	if !bytes.Equal(randaoReveal[:], epochSignature[:]) {
		t.Errorf("Expected randao reveals to be equal, received %#x != %#x", randaoReveal[:], epochSignature[:])
	}
}
