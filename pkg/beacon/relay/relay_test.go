package relay

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/keep-network/keep-core/pkg/beacon/relay/config"
	chainLocal "github.com/keep-network/keep-core/pkg/chain/local"
)

var address = "0x65ea55c1f10491038425725dc00dffeab2a1e28a"
var relayEntryTimeout = uint64(15)

func TestMonitorRelayEntryOnChain_EntrySubmitted(t *testing.T) {
	chain := chainLocal.Connect(5, 3, big.NewInt(200))
	blockCounter, err := chain.BlockCounter()
	if err != nil {
		fmt.Printf("failed to setup a block counter: [%v]", err)
	}

	node := &Node{
		blockCounter: blockCounter,
	}

	relayChain := chain.ThresholdRelay()
	chainConfig := &config.Chain{
		RelayEntryTimeout: uint64(relayEntryTimeout),
	}
	startBlockHeight, err := blockCounter.CurrentBlock()
	if err != nil {
		t.Fatal(err)
	}

	go node.MonitorRelayEntry(
		relayChain,
		startBlockHeight,
		chainConfig,
	)

	// the window to get a relay entry is from currentBlock to (currentBlock+relayEntryTimeout)
	// we subtract arbitarly 5 blocks to be within this window. Ex. 0 + 15 - 5
	relayEntryResultWindow := startBlockHeight + relayEntryTimeout - 5
	err = blockCounter.WaitForBlockHeight(relayEntryResultWindow)
	if err != nil {
		fmt.Printf("failed to wait for a block: [%v]. Error occured: [%v]", relayEntryResultWindow, err)
	}

	chain.ThresholdRelay().SubmitRelayEntry(big.NewInt(1)).OnFailure(func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	})

	blockCounter.WaitForBlockHeight(startBlockHeight + relayEntryTimeout)

	timeoutsReport := chain.GetReportRelayEntryTimeouts()
	numberOfReports := len(timeoutsReport)

	if numberOfReports != 0 {
		t.Fatalf(
			"\nexpected: [%v]\nactual:   [%v]",
			0,
			numberOfReports,
		)
	}
}

func TestMonitorRelayEntryOnChain_EntryNotSubmitted(t *testing.T) {
	chain := chainLocal.Connect(5, 3, big.NewInt(200))
	blockCounter, err := chain.BlockCounter()
	if err != nil {
		fmt.Printf("failed to setup a block counter: [%v]", err)
	}

	node := &Node{
		blockCounter: blockCounter,
	}

	relayChain := chain.ThresholdRelay()
	chainConfig := &config.Chain{
		RelayEntryTimeout: uint64(relayEntryTimeout),
	}
	currentBlock, err := blockCounter.CurrentBlock()
	if err != nil {
		t.Fatal(err)
	}

	go node.MonitorRelayEntry(
		relayChain,
		currentBlock,
		chainConfig,
	)

	// we want to exceed the relay entry timeout to report that a relay entry
	// was not submitted. 5 is an arbitrary number to exceed relayEntryTimeout.
	blockCounter.WaitForBlockHeight(currentBlock + relayEntryTimeout + 5)

	timeoutsReport := chain.GetReportRelayEntryTimeouts()
	numberOfReports := len(timeoutsReport)

	if numberOfReports != 1 {
		t.Fatalf(
			"Number of timeout reports does not match\nexpected: [%v]\nactual:   [%v]",
			1,
			numberOfReports,
		)
	}

	if timeoutsReport[0] != relayEntryTimeout {
		t.Fatalf(
			"Timeout reporting must happen only after a relay entry timeout\nexpected: [%v]\nactual:   [%v]",
			relayEntryTimeout,
			timeoutsReport[0],
		)
	}
}
