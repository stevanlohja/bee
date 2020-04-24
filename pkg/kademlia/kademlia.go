// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kademlia

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethersphere/bee/pkg/addressbook"
	"github.com/ethersphere/bee/pkg/discovery"
	"github.com/ethersphere/bee/pkg/kademlia/pslice"
	"github.com/ethersphere/bee/pkg/logging"
	"github.com/ethersphere/bee/pkg/p2p"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/ethersphere/bee/pkg/topology"
)

const (
	maxBins        = 16
	nnLowWatermark = 2 // the number of peers in consecutive deepest bins that constitute as nearest neighbours
)

type Kad struct {
	base           swarm.Address // this node's overlay address
	discovery      discovery.Driver
	addressBook    addressbook.Interface
	p2p            p2p.Service
	connectedPeers *pslice.PSlice // a slice of peers sorted and indexed by po, indexes kept in `bins`
	knownPeers     *pslice.PSlice // both are po aware slice of addresses
	depth          uint8          // current neighborhood depth
	mtx            sync.RWMutex

	manageC chan struct{}

	logger logging.Logger

	quit chan struct{}
}

func New(base swarm.Address, disc discovery.Driver, addressBook addressbook.Interface, p2pService p2p.Service, logger logging.Logger) *Kad {
	k := &Kad{
		base:        base,
		discovery:   disc,
		addressBook: addressBook,
		p2p:         p2pService,

		connectedPeers: pslice.New(maxBins),
		knownPeers:     pslice.New(maxBins),

		manageC: make(chan struct{}, 1),
		logger:  logger,
		quit:    make(chan struct{}),
	}

	go k.manage()
	return k
}

// manage is a forever loop that manages the connection to new peers
// once they get added or once others leave.

func (k *Kad) manage() {
	for {
		select {
		case <-k.quit:
			fmt.Println("*********shutting down*****")
			return
		case <-k.manageC:
			// this gets triggered every time there's a change in the information about
			// the peer we know about. a peer could have been removed (i.e. disconnected)
			// or added (from discovery, debugapi, bootnode flag or just because she was
			// persisted in the address book and the node has been restarted)

			// so now we need to manage the change, but do so carefully so we dont end up
			// connecting to everyone we know

			// a hard rule would say that forever, every peer we know about which has PO >= d, we should connect to that peer
			// this should always be enforced and potentially is the first part of the iteration here

			// at bootstrapping though, this could end up with us connecting to all peers since they all fit the criteria
			// so a mandatory mark-and-sweep should be done after each connection is made

			// a more robust approach would be:
			// establishing connections consists of two phases:
			// - shallower-than-depth
			// - deeper | greater than depth

			// for the first part, the strategy is always to make the depth go deeper
			// this means we have to consistently have peers in all bins from bin 0, going on deeper up to max bins
			// also, empty bins should be mitigated as they automatically constitute as depth when they are a "hole" in the
			// bins (one bin shallower than others with no peers)
			// when connecting to a new peer depth has to be recalculated, if the given lower threshold for a bin has been reached,
			// we simply iterate to the next one, connecting as necessary, until we reach depth

			// when we iterate on peers which have po > depth, we should connect to all of them,
			// but this should be done in an iterative manner, since they could potentially make depth to go even deeper
			// it could be that a way to measure this would be that when we connect to a new peer, the depth does not go any deeper (or every 2 peers)
			// and this means we can just blindly connect to anything which is deeper than depth (???)

			// strategy: push depth down (start shallow and go deeper)
			// want to connect, for example, to two peers in a bin
			// if i have more - dont want to connect(but others in the bin might dial in; TODO high watermark to remove peers with mark & sweep)
			// this is (potentially) outside of depth so start from outside of depth
			// and start moving depth deeper, once there's no more peers/bins we'd like to
			// connect to peers to outside of depth, connect to peers from inside depth

			_ = k.knownPeers.EachBinRev(func(peer swarm.Address, po uint8) (bool, bool, error) {
				//fmt.Println("peer ", peer.String(), " po ", po)

				if k.connectedPeers.Exists(peer) {
					return false, false, nil
				}

				if k.binSaturated(po, k.depth) {
					return false, true, nil
				}

				// connect to peers in the bin
				ma, err := k.addressBook.Get(peer)
				if err != nil {
					// do something and remove the entry from the slice
					panic("err")
					return false, false, nil //go to next peer
				}
				k.logger.Debugf("kademlia dialing to peer %s", peer.String())
				_, err = k.p2p.Connect(context.Background(), ma)
				if err != nil {
					k.logger.Debugf("error connecting to peer %s: %v", peer, err)
					// TODO: somehow keep track of attempts and at some point forget about the peer
					return false, false, nil // dont stop, continue to next peer
				}

				k.mtx.Lock()
				k.logger.Debugf("connected to peer: %s", peer)
				k.connectedPeers.Add(peer, po)
				k.depth = k.recalcDepth()
				k.mtx.Unlock()
				select {
				case k.manageC <- struct{}{}:
				case <-k.quit:
				default:
				}

				// at this point we dont know if the bin is saturated or not
				// or whether there are any other peers to connect to in that bin
				// so lets continue to the next peer in the bin and do the checks again
				//return false, false, nil
				return true, false, nil
			})
		}
	}
}

// binSaturated indicates whether a certain bin is saturated or not.
// when a bin is not saturated it means we would like to proactively
// initiate connections to other peers in the bin.
func (k *Kad) binSaturated(po, depth uint8) bool {
	if po >= depth {
		// short circuit for bins which are >= depth
		return false
	}

	return false
	//var end int
	//if po == maxBins-1 {
	//end = len(k.peers)
	//} else {
	//end = int(s.bins[po+1])
	//}

	//size := end - int(s.bins[po])
	size := 0

	// lets assume for now that the minimum number of peers in a bin
	// would be 2, under which we would always want to connect to new peers
	// obviously this should be replaced with a better optimization
	val := 2
	if size < val {
		return false
	}
	return true
}

// recalcDepth returns the kademlia depth. Must be called under lock.
func (k *Kad) recalcDepth() uint8 {
	// start from deepest
	// find minNNPeers and set as candidate
	// check no empty intermediate bins
	// if no empty bins - set candidate as depth
	// if empty bins - set depth as depth of shallowest empty bin

	// handle edge case separately
	if k.connectedPeers.Length() <= nnLowWatermark {
		return 0
	}
	var (
		peers                        = uint(0)
		candidate                    = uint8(0)
		shallowestEmpty, noEmptyBins = k.connectedPeers.ShallowestEmpty()
	)

	_ = k.connectedPeers.EachBin(func(_ swarm.Address, po uint8) (bool, bool, error) {
		peers++
		if peers >= nnLowWatermark {
			candidate = po
			return true, false, nil
		}
		return false, false, nil
	})

	if noEmptyBins || shallowestEmpty > candidate {
		return candidate
	}

	return shallowestEmpty
}

func (k *Kad) AddPeer(ctx context.Context, addr swarm.Address) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()

	var (
		e  = k.connectedPeers.Exists(addr)
		e2 = k.knownPeers.Exists(addr)
	)

	if e || e2 {
		// TODO return a custom error here or change the method signature to return a bool too
		panic("peer already connected")
	}

	// both false
	po := swarm.Proximity(k.base.Bytes(), addr.Bytes())
	k.knownPeers.Add(addr, uint8(po))
	// add to address book too???

	select {
	case k.manageC <- struct{}{}:
	case <-k.quit:
	default:
	}

	return nil
}

// called when peer disconnects
func (k *Kad) Disconnected(addr swarm.Address) {
	k.mtx.Lock()
	defer k.mtx.Unlock()

	po := uint8(swarm.Proximity(k.base.Bytes(), addr.Bytes()))
	k.connectedPeers.Remove(addr, po)
	k.depth = k.recalcDepth()

	select {
	case k.manageC <- struct{}{}:
	case <-k.quit:
	default:
	}
}

func (k *Kad) Remove(addr swarm.Address) {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	po := swarm.Proximity(k.base.Bytes(), addr.Bytes())
	k.connectedPeers.Remove(addr, uint8(po))
	k.depth = k.recalcDepth()

	select {
	case k.manageC <- struct{}{}:
	case <-k.quit:
	default:
	}
}

func (k *Kad) ClosestPeer(addr swarm.Address) (peerAddr swarm.Address, err error) {
	panic("not implemented") // TODO: Implement
}

func (k *Kad) EachPeerScored(addr swarm.Address, scoreF topology.ScoreFunc, pf topology.EachPeerFunc) error {
	//po := po(k.base, peer)

	//for i := po; i >= 0; i-- {
	//scores := []scoredPeer{}
	//peers := k.peers[k.bins[po]:k.bins[po+1]]

	//// score all peers in the bin
	//for _, p := range peers {
	//sp := scoredPeer{
	//addr:  p,
	//score: scoreF(p),
	//}
	//}

	//// sort.ByScore(peers)...

	//// call pf with each peer
	//for _, v := range scores {
	//stop, err := pf(v.addr)
	//// if stop then exit, if err then exit etc...
	//}
	//}

	return nil
}

// NeighborhoodDepth returns the current Kademlia depth.
func (k *Kad) NeighborhoodDepth() uint8 {
	k.mtx.Lock()
	defer k.mtx.Unlock()

	return k.neighborhoodDepth()
}

func (k *Kad) neighborhoodDepth() uint8 {
	return k.depth
}

func (k *Kad) MarshalJSON() ([]byte, error) {
	k.mtx.Lock()
	defer k.mtx.Unlock()

	// which kind of data we want to see?
	// each bin size (how many peers we are connected to)
	// each bin potential (how many peers we know of but not connected to)
	// in the future maybe some coefficient about address space coverage in a particular bin
	// total population (aggregate size) - both connected and not connected
	// this node's address
	// timestamp, maxBins, nnLowWatermark, current depth, peers in each bin

	var (
	//size = 0 // uint(len(k.peers))
	)

	return []byte(""), nil
	//type binInfo struct {
	//BinSize        uint
	//BinKnown       uint
	//BinConnected   uint
	//KnownPeers     []string
	//ConnectedPeers []string
	//}

	//type kadBins struct {
	//Bin0  binInfo `json:"bin_0"`
	//Bin1  binInfo `json:"bin_1"`
	//Bin2  binInfo `json:"bin_2"`
	//Bin3  binInfo `json:"bin_3"`
	//Bin4  binInfo `json:"bin_4"`
	//Bin5  binInfo `json:"bin_5"`
	//Bin6  binInfo `json:"bin_6"`
	//Bin7  binInfo `json:"bin_7"`
	//Bin8  binInfo `json:"bin_8"`
	//Bin9  binInfo `json:"bin_9"`
	//Bin10 binInfo `json:"bin_10"`
	//Bin11 binInfo `json:"bin_11"`
	//Bin12 binInfo `json:"bin_12"`
	//Bin13 binInfo `json:"bin_13"`
	//Bin14 binInfo `json:"bin_14"`
	//Bin15 binInfo `json:"bin_15"`
	//}

	//type kadParams struct {
	//Base           string //base address string
	//Population     int    //known
	//Connected      int    //connected count
	//Timestamp      time.Time
	//NNLowWatermark int
	//Depth          uint8
	//Bins           kadBins
	//}

	//infos := make([]binInfo, maxBins)

	//// TODO consider to switch to eachBin but it might get a bit ugly
	//for i := (maxBins - 1); i >= 0; i-- {
	//binSize := size - k.bins[i]
	//connected := k.peers[k.bins[i] : k.bins[i]+binSize]
	//var connectedPeers []string
	//for _, v := range connected {
	//connectedPeers = append(connectedPeers, v.String())
	//}

	//infos[i] = binInfo{
	//BinSize:        binSize, // should be binKnown+binConnected
	//BinKnown:       0,
	//BinConnected:   binSize,
	//KnownPeers:     []string{}, //still todo
	//ConnectedPeers: connectedPeers,
	//}

	//size = k.bins[i]
	//}

	//return json.MarshalIndent(kadParams{
	//Base:           k.base.String(),
	//Population:     len(k.peers), //still todo
	//Connected:      len(k.peers),
	//Timestamp:      time.Now(),
	//NNLowWatermark: nnLowWatermark,
	//Depth:          k.depth,
	//Bins: kadBins{
	//Bin0:  infos[0],
	//Bin1:  infos[1],
	//Bin2:  infos[2],
	//Bin3:  infos[3],
	//Bin4:  infos[4],
	//Bin5:  infos[5],
	//Bin6:  infos[6],
	//Bin7:  infos[7],
	//Bin8:  infos[8],
	//Bin9:  infos[9],
	//Bin10: infos[10],
	//Bin11: infos[11],
	//Bin12: infos[12],
	//Bin13: infos[13],
	//Bin14: infos[14],
	//Bin15: infos[15],
	//},
	//}, "", "  ")
}

func (k *Kad) String() string {
	b, _ := k.MarshalJSON()
	return string(b)
}

func (k *Kad) Close() error {
	close(k.quit)
	return nil
}

type scoredPeer struct {
	addr  swarm.Address
	score float32
}
