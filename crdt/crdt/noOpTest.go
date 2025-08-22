package crdt

import (
	"errors"
	"fmt"
	"potionDB/crdt/clocksi"
)

type ReplicaNoOp struct {
	Crdt *NoOpCrdt
	Id   int16
}

type ReplicaOp struct {
	Op         Operation
	ReplicaNum int
}

type ReplicaMessage struct {
	Msg        Message
	ReplicaNum int
	Clock      clocksi.ClockSiTimestamp
}

const ID_BASE int16 = 11

var (
	addArtistSamRep1 = ReplicaOp{&AddArtist{ArtistName: "Sam"}, 1}
	addArtistSamRep3 = ReplicaOp{&AddArtist{ArtistName: "Sam"}, 3}
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
	var opOrder = [][]ReplicaOp{{addArtistSamRep1}, {addAlbum1Rep1, addAlbum2Rep2}, {updArtistSamRep1, rmvArtistSamRep2}}
	err := testReplicas(2, opOrder)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestNoOpCrdt2() {
	fmt.Println("\n Test start TestNoOpCrdt2")
	var opOrder = [][]ReplicaOp{{addArtistSamRep1}, {addAlbum1Rep1, addAlbum2Rep2}, {rmvArtistFredRep2, updArtistSamRep1}}
	err := testReplicas(2, opOrder)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestNoOpCrdt3() {
	fmt.Println("\n Test start TestNoOpCrdt3")
	var opOrder = [][]ReplicaOp{{addArtistSamRep1}, {addAlbum1Rep1, addAlbum2Rep2}, {rmvArtistFredRep2, updArtistSamRep1}, {addArtistFredRep2}, {addAlbumFredRep2}}
	err := testReplicas(2, opOrder)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestNoOpCrdt4() {
	fmt.Println("\n Test start TestNoOpCrdt4")
	var opOrderR1 = [][]ReplicaOp{{addArtistSamRep1, addArtistFredRep2}, {addAlbum1Rep1, addAlbum2Rep2}, {rmvArtistFredRep2, updArtistSamRep1}, {addAlbumFredRep2}}
	err := testReplicas(2, opOrderR1)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestNoOpCrdt5() {
	fmt.Println("\n Test start TestNoOpCrdt5")
	var opOrderR1 = [][]ReplicaOp{{addArtistSamRep3, addAlbum1Rep1, addAlbum2Rep2}}
	err := testReplicas(3, opOrderR1)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Func that carries out basic test structure
// base is the base state of all replicas
func testReplicas(numReplicas int, opBlocks [][]ReplicaOp) error {

	if numReplicas > 32767 || numReplicas <= 0 {
		return errors.New("NoOpTest parameter error: Invalid number of replicas")
	} else if numReplicas*int(ID_BASE) > 32767 {
		return errors.New("NoOpTest parameter error: Invalid number of replicas for value of base id")
	}
	reps := make([]*ReplicaNoOp, numReplicas)
	//Creating replicas - MusicState is assumed for use
	for i := 0; i < numReplicas; i++ {
		var newId int16 = ID_BASE * (int16(i + 1))
		crdt := (&NoOpCrdt{}).Initialize(nil, newId).(*NoOpCrdt)
		crdt.StateContent = &MusicState{}
		crdt.StateContent.Initialize()
		reps[i] = &ReplicaNoOp{Crdt: crdt, Id: newId}
	}

	fmt.Printf("Starting %v Replica(s).\n", numReplicas)
	processOps(reps, opBlocks)
	for i, rep := range reps {
		fmt.Printf("Final result, Replica%v:\n", i+1)
		PrintNoOpCrdt(rep.Crdt, fmt.Sprintf("Replica%v", i+1))
	}
	return nil
}

// Func that prepares and builds the graph in the crdt taking into account operation conflicts
func processOps(reps []*ReplicaNoOp, opBlocks [][]ReplicaOp) error {
	i := 0
	//Hard-Initialization of the clock. using .NewTimestamp had
	vectorClock := make(map[int16]int64)
	for _, rep := range reps {
		vectorClock[rep.Id] = 0
	}

	var clock clocksi.ClockSiTimestamp = clocksi.ClockSiTimestamp{VectorClock: vectorClock}
	for _, opBlock := range opBlocks {
		conflictMsgs := []ReplicaMessage{}
		j := i
		clockCpy := clock.Copy().(clocksi.ClockSiTimestamp) //performs changes from current state.
		for _, op := range opBlock {                        //Prepare messages for all conflicting messages and apply to source replica before replicating downstream
			if op.ReplicaNum > len(reps) {
				return errors.New("NoOpTest operation error: Target Replica out of bounds")
			}
			msg := reps[op.ReplicaNum-1].Crdt.Update(op.Op).(Message)
			clock.SelfIncTimestamp(reps[op.ReplicaNum-1].Id) //keeps track of total updates
			nextClock := clockCpy.IncTimestamp(reps[op.ReplicaNum-1].Id).(clocksi.ClockSiTimestamp)
			conflictMsgs = append(conflictMsgs, ReplicaMessage{Msg: msg, ReplicaNum: op.ReplicaNum, Clock: nextClock})
			reps[op.ReplicaNum-1].Crdt.Downstream(nextClock, msg)
			j++
		}
		/*
			Replicate downstream
		*/
		for k := i; k < j; k++ {
			msgWithSourceRep := conflictMsgs[k-i]
			srcRep := msgWithSourceRep.ReplicaNum
			for l := range reps {
				if l+1 != srcRep {
					reps[l].Crdt.Downstream(msgWithSourceRep.Clock, msgWithSourceRep.Msg)
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
	fmt.Println("EndState.")
}

// Assumes the format is correct, only for checking test output and debugging
func PrintMusicStateFromBytes(stateByte [][]byte) {
	PrintMusicState((&MusicState{}).Deserialize(stateByte).(*MusicState))
}
