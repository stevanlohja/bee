// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kademlia_test

import (
	"context"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
	"time"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/ethersphere/bee/pkg/addressbook"
	"github.com/ethersphere/bee/pkg/kademlia"
	"github.com/ethersphere/bee/pkg/logging"
	p2pmock "github.com/ethersphere/bee/pkg/p2p/mock"
	mockstate "github.com/ethersphere/bee/pkg/statestore/mock"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/ethersphere/bee/pkg/topology/test"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestNeighborhoodDepth(t *testing.T) {
	var (
		base   = test.RandomAddress()                       // base address
		logger = logging.New(os.Stdout, 3)                  // logger
		ab     = addressbook.New(mockstate.NewStateStore()) // address book
		conns  int32                                        // how many connect calls were made to the p2p mock

		p2p = p2pmock.New(p2pmock.WithConnectFunc(func(_ context.Context, addr ma.Multiaddr) (swarm.Address, error) {
			_ = atomic.AddInt32(&conns, 1)
			return swarm.ZeroAddress, nil
		}))

		kad      = kademlia.New(base, nil, ab, p2p, logger)
		peers    []swarm.Address
		binEight []swarm.Address
	)

	defer kad.Close()

	for i := 0; i < 8; i++ {
		addr := test.RandomAddressAt(base, i)
		peers = append(peers, addr)
	}

	for i := 0; i < 2; i++ {
		addr := test.RandomAddressAt(base, 8)
		binEight = append(binEight, addr)
	}

	// check empty kademlia depth is 0
	kDepth(t, kad, 0)

	// add two bin 8 peers, verify depth still 0
	add(t, kad, ab, binEight, 0, 2)
	kDepth(t, kad, 0)

	// add two first peers (po0,po1)
	add(t, kad, ab, peers, 0, 2)

	// wait for 4 connections
	waitConns(t, &conns, 4)

	// depth 2 (shallowest empty bin)
	kDepth(t, kad, 2)

	// reset the counter
	atomic.StoreInt32(&conns, 0)

	for i := 2; i < len(peers)-1; i++ {
		addOne(t, kad, ab, peers[i])

		// wait for one connection
		waitConn(t, &conns)

		// depth is i+1
		kDepth(t, kad, i+1)

		// reset
		atomic.StoreInt32(&conns, 0)
	}

	// the last peer in bin 7 which is empty we insert manually,
	addOne(t, kad, ab, peers[len(peers)-1])
	waitConn(t, &conns)

	// depth is 8 because we have nnLowWatermark neighbors in bin 8
	kDepth(t, kad, 8)

	atomic.StoreInt32(&conns, 0)

	// now add another ONE peer at depth+1, and expect the depth to still
	// stay 8, because the counter for nnLowWatermark would be reached only at the next
	// depth iteration when calculating depth
	addr := test.RandomAddressAt(base, 9)
	addOne(t, kad, ab, addr)
	waitConn(t, &conns)
	kDepth(t, kad, 8)

	atomic.StoreInt32(&conns, 0)

	// fill the rest up to the last bin and check that everything works at the edges
	for i := 10; i < kademlia.MaxBins; i++ {
		addr := test.RandomAddressAt(base, i)
		addOne(t, kad, ab, addr)
		waitConn(t, &conns)
		kDepth(t, kad, i-1)
		atomic.StoreInt32(&conns, 0)
	}

	// add one more at the end, and verify depth 15
	addr = test.RandomAddressAt(base, 15)
	addOne(t, kad, ab, addr)
	waitConn(t, &conns)
	kDepth(t, kad, 15)

	// now remove that peer and check that the depth is back at 14
	removeOne(kad, addr)
	kDepth(t, kad, 14)

	// remove the peer at bin 1, depth should be 1
	removeOne(kad, peers[1])
	kDepth(t, kad, 1)
}

func removeOne(k *kademlia.Kad, peer swarm.Address) {
	k.Remove(peer)
}

const underlayBase = "/ip4/127.0.0.1/tcp/7070/dns/"

func addOne(t *testing.T, k *kademlia.Kad, ab addressbook.Putter, peer swarm.Address) {
	t.Helper()
	multiaddr, err := ma.NewMultiaddr(underlayBase + peer.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := ab.Put(peer, multiaddr); err != nil {
		t.Fatal(err)
	}
	k.AddPeer(nil, peer)
}

func add(t *testing.T, k *kademlia.Kad, ab addressbook.Putter, peers []swarm.Address, offset, number int) {
	t.Helper()
	for i := offset; i < offset+number; i++ {
		multiaddr, err := ma.NewMultiaddr(underlayBase + peers[i].String())
		if err != nil {
			t.Fatal(err)
		}
		if err := ab.Put(peers[i], multiaddr); err != nil {
			t.Fatal(err)
		}
		k.AddPeer(nil, peers[i])
	}
}

func kDepth(t *testing.T, k *kademlia.Kad, d int) {
	t.Helper()
	for i := 0; i < 20; i++ {
		depth := int(k.NeighborhoodDepth())
		if depth == d {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for depth. want %d", d)
}

func waitConn(t *testing.T, conns *int32) {
	t.Helper()
	waitConns(t, conns, 1)
}

func waitConns(t *testing.T, conns *int32, exp int32) {
	t.Helper()
	var got int32
	for i := 0; i < 50; i++ {
		got = atomic.LoadInt32(conns)
		if got == exp {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for connections to be established. got %d want %d", got, exp)
}
