package crdt

import (
	"errors"
	"fmt"
	"potionDB/crdt/clocksi"
)

type ReplicaOp struct {
	Op Operation
	ReplicaNum int
}

type ReplicaMessage struct {
	Msg Message
	ReplicaNum int
}

var (
	addArtistSamRep1 = ReplicaOp{&AddArtist{ArtistName: "Sam"}, 1}
	addAlbum1Rep1    = ReplicaOp{&AddAlbum{AlbumName: "A1", ArtistName: "Sam"}, 1}
	addAlbum2Rep2    = ReplicaOp{&AddAlbum{AlbumName: "A2", ArtistName: "Sam"}, 2}
	updArtistSamRep1 = ReplicaOp{&UpdArtist{ArtistName: "Sam"}, 1}
	rmvArtistSamRep2 = ReplicaOp{&RmvArtist{ArtistName: "Sam"}, 2}

	addArtistFredRep2 = ReplicaOp{&AddArtist{ArtistName: "Fred"}, 2}
	addAlbumFredRep2  = ReplicaOp{&AddAlbum{AlbumName: "The Great Pretender", ArtistName: "Fred"}, 2}
	rmvArtistFredRep2 = ReplicaOp{&RmvArtist{ArtistName: "Fred"}, 2}
)

func TestNoOpCrdt1() {
	fmt.Println("\n Test start TestNoOpCrdt1")

	crdt := (&NoOpCrdt{}).Initialize(nil, 111).(*NoOpCrdt)
	crdt.StateContent = &MusicState{}
	crdt.StateContent.Initialize()
	var opOrder = [][]ReplicaOp{{addArtistSamRep1}, {addAlbum1Rep1, addAlbum2Rep2}, {updArtistSamRep1, rmvArtistSamRep2}}
	//create specific timestamps
	var timestamps = []clocksi.ClockSiTimestamp{
		{VectorClock: map[int16]int64{111: 1, 222: 0}},
		{VectorClock: map[int16]int64{111: 2, 222: 0}},
		{VectorClock: map[int16]int64{111: 1, 222: 1}},
		{VectorClock: map[int16]int64{111: 3, 222: 1}},
		{VectorClock: map[int16]int64{111: 2, 222: 2}}}

	err := testReplicas(crdt, 2, opOrder, timestamps)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestNoOpCrdt2() {
	fmt.Println("\n Test start TestNoOpCrdt2")

	crdt := (&NoOpCrdt{}).Initialize(nil, 111).(*NoOpCrdt)
	crdt.StateContent = &MusicState{}
	crdt.StateContent.Initialize()
	var opOrder = [][]ReplicaOp{{addArtistSamRep1}, {addAlbum1Rep1, addAlbum2Rep2}, {rmvArtistFredRep2, updArtistSamRep1}}
	//create specific timestamps
	var timestamps = []clocksi.ClockSiTimestamp{
		{VectorClock: map[int16]int64{111: 1, 222: 0}},
		{VectorClock: map[int16]int64{111: 2, 222: 0}},
		{VectorClock: map[int16]int64{111: 1, 222: 1}},
		{VectorClock: map[int16]int64{111: 3, 222: 1}},
		{VectorClock: map[int16]int64{111: 2, 222: 2}}}

	err := testReplicas(crdt, 2, opOrder, timestamps)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestNoOpCrdt3() {
	fmt.Println("\n Test start TestNoOpCrdt3")

	crdt := (&NoOpCrdt{}).Initialize(nil, 111).(*NoOpCrdt)
	crdt.StateContent = &MusicState{}
	crdt.StateContent.Initialize()
	var opOrder = [][]ReplicaOp{{addArtistSamRep1}, {addAlbum1Rep1, addAlbum2Rep2}, {rmvArtistFredRep2, updArtistSamRep1}, {addArtistFredRep2}, {addAlbumFredRep2}}
	//create specific timestamps
	var timestamps = []clocksi.ClockSiTimestamp{
		{VectorClock: map[int16]int64{111: 1, 222: 0}},
		{VectorClock: map[int16]int64{111: 2, 222: 0}},
		{VectorClock: map[int16]int64{111: 1, 222: 1}},
		{VectorClock: map[int16]int64{111: 3, 222: 1}},
		{VectorClock: map[int16]int64{111: 2, 222: 2}},
		{VectorClock: map[int16]int64{111: 3, 222: 2}},
		{VectorClock: map[int16]int64{111: 4, 222: 2}}}

	err := testReplicas(crdt, 2, opOrder, timestamps)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestNoOpCrdt4() {
	fmt.Println("\n Test start TestNoOpCrdt4")

	crdt := (&NoOpCrdt{}).Initialize(nil, 111).(*NoOpCrdt)
	crdt.StateContent = &MusicState{}
	crdt.StateContent.Initialize()
	var opOrderR1 = [][]ReplicaOp{{addArtistSamRep1, addArtistFredRep2}, {addAlbum1Rep1, addAlbum2Rep2}, {rmvArtistFredRep2, updArtistSamRep1}, {addAlbumFredRep2}}
	//create specific timestamps
	var timestamps = []clocksi.ClockSiTimestamp{
		{VectorClock: map[int16]int64{111: 1, 222: 0}},
		{VectorClock: map[int16]int64{111: 0, 222: 1}},
		{VectorClock: map[int16]int64{111: 2, 222: 1}},
		{VectorClock: map[int16]int64{111: 1, 222: 2}},
		{VectorClock: map[int16]int64{111: 3, 222: 2}},
		{VectorClock: map[int16]int64{111: 2, 222: 3}},
		{VectorClock: map[int16]int64{111: 4, 222: 4}}}

	err := testReplicas(crdt, 2, opOrderR1, timestamps)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Func that carries out basic test structure
// base is the base state of all replicas
func testReplicas(base *NoOpCrdt, numReplicas int, opBlocks [][]ReplicaOp, timestamps []clocksi.ClockSiTimestamp) error {
	reps := make([]*NoOpCrdt, numReplicas)
	if numReplicas > 32767 || numReplicas <= 0{
		return errors.New("NoOpTest parameter error: Invalid number of replicas")
	}
	//Creating replicas from base crdt
	for i := 0; i < numReplicas; i++ {
		reps[i] = base.Copy().(*NoOpCrdt)
	}
	fmt.Printf("Starting %v Replica(s).\n", numReplicas)
	processOps(reps, opBlocks, timestamps)
	for i, crdt := range reps {
		fmt.Printf("Final result, Replica%v:\n", i+1)
		PrintNoOpCrdt(crdt, fmt.Sprintf("Replica%v", i+1))
	}
	return nil
}

// Func that prepares and builds the graph in the crdt taking into account operation conflicts
func processOps(reps []*NoOpCrdt, opBlocks [][]ReplicaOp, timestamps []clocksi.ClockSiTimestamp) error {
	i := 0
	for _, opBlock := range opBlocks {
		conflictMsgs := []ReplicaMessage{}
		j := i
		for _, op := range opBlock { //Prepare messages for all conflicting messages and apply to source replica before replicating downstream
			if (op.ReplicaNum > len(reps)) {
				return errors.New("NoOpTest operation error: Target Replica out of bounds")
			}
			msg := reps[op.ReplicaNum-1].Update(op.Op).(Message)
			conflictMsgs = append(conflictMsgs, ReplicaMessage{Msg: msg, ReplicaNum: op.ReplicaNum})
			reps[op.ReplicaNum-1].Downstream(timestamps[j], msg)
			j++
		}
		/*
		Replicate downstream
		*/
		for k := i; k < j; k++ {
			msgWithSourceRep :=  conflictMsgs[k-i]
			srcRep := msgWithSourceRep.ReplicaNum
			msg := msgWithSourceRep.Msg
			for l := range reps {
				if (l+1 != srcRep) {
					reps[l].Downstream(timestamps[k], msg)
				}
			}
		}
		i = j
	}
	return nil
}

// Func that prints out contents of a NoOpCrdt
func PrintNoOpCrdt(c *NoOpCrdt, name string) {
	fmt.Println("--------------------", name)
	fmt.Println("Operations:")
	for _, node := range c.NodeArr {
		fmt.Println(node.Value.Op, node.IsNoOp)
		//print clock
		conv := node.Value.Clock.(clocksi.ClockSiTimestamp)
		for key, val := range conv.VectorClock {
			fmt.Printf("[%v : %v] ", key, val)
		}
		fmt.Printf("\n")
		/*
			fmt.Println("Blocks----")
			for _, op := range node.Value.BlockedOps {
				fmt.Println("--", op)
			}
			fmt.Println("--------")
		*/
	}
	PrintMusicState(c.Read(nil, nil).(*MusicState))
}

// Func that prints out the data (Music State -> Artists and Albums) for a NoOpCrdt using MusicState
func PrintMusicState(d *MusicState) {
	fmt.Println("State:")
	for artist, val := range d.State {
		fmt.Println("Artist:", artist)
		fmt.Println("-- Albums: ", val.Len())
		for _, album := range val.Keys() {
			fmt.Println("----", album)
		}
	}
}

//Assumes the format is correct, only for checking test output and debugging
func PrintMusicStateFromBytes(stateByte [][]byte) {
	PrintMusicState((&MusicState{}).Deserialize(stateByte).(*MusicState))
}
