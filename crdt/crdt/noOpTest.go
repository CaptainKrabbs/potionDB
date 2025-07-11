package crdt

import (
	"fmt"
	"potionDB/crdt/clocksi"
)

var (
	addArtistSam = AddArtist{ArtistName: "Sam"}
	addAlbum1    = AddAlbum{AlbumName: "A1", ArtistName: "Sam"}
	addAlbum2    = AddAlbum{AlbumName: "A2", ArtistName: "Sam"}
	updArtistSam = UpdArtist{ArtistName: "Sam"}
	rmvArtistSam = RmvArtist{ArtistName: "Sam"}

	addArtistFred = AddArtist{ArtistName: "Fred"}
	addAlbumFred  = AddAlbum{AlbumName: "The Great Pretender", ArtistName: "Fred"}
	rmvArtistFred = RmvArtist{ArtistName: "Fred"}
)

func TestNoOpCrdt1() {
	fmt.Println("\n Test start TestNoOpCrdt1")
	crdtR1 := (&NoOpCrdt{}).Initialize(nil, 111).(*NoOpCrdt)
	crdtR2 := (&NoOpCrdt{}).Initialize(nil, 222).(*NoOpCrdt)

	var opOrderR1 = [][]Operation{{&addArtistSam}, {&addAlbum1, &addAlbum2}, {&updArtistSam, &rmvArtistSam}}
	var opOrderR2 = [][]Operation{{&addArtistSam}, {&addAlbum1, &addAlbum2}, {&rmvArtistSam, &updArtistSam}}
	//create specific timestamps
	var timestamps = []clocksi.ClockSiTimestamp{
		{VectorClock: map[int16]int64{111: 1, 222: 0}},
		{VectorClock: map[int16]int64{111: 2, 222: 0}},
		{VectorClock: map[int16]int64{111: 1, 222: 1}},
		{VectorClock: map[int16]int64{111: 3, 222: 1}},
		{VectorClock: map[int16]int64{111: 2, 222: 2}}}

	testReplica(crdtR1, 1, opOrderR1, timestamps)
	testReplica(crdtR2, 2, opOrderR2, timestamps)
}

func TestNoOpCrdt2() {
	fmt.Println("\n Test start TestNoOpCrdt2")
	crdtR1 := (&NoOpCrdt{}).Initialize(nil, 111).(*NoOpCrdt)

	var opOrderR1 = [][]Operation{{&addArtistSam}, {&addAlbum1, &addAlbum2}, {&rmvArtistFred, &updArtistSam}}
	//create specific timestamps
	var timestamps = []clocksi.ClockSiTimestamp{
		{VectorClock: map[int16]int64{111: 1, 222: 0}},
		{VectorClock: map[int16]int64{111: 2, 222: 0}},
		{VectorClock: map[int16]int64{111: 1, 222: 1}},
		{VectorClock: map[int16]int64{111: 3, 222: 1}},
		{VectorClock: map[int16]int64{111: 2, 222: 2}}}

	testReplica(crdtR1, 1, opOrderR1, timestamps)
}

func TestNoOpCrdt3() {
	fmt.Println("\n Test start TestNoOpCrdt3")
	crdtR1 := (&NoOpCrdt{}).Initialize(nil, 111).(*NoOpCrdt)

	var opOrderR1 = [][]Operation{{&addArtistSam}, {&addAlbum1, &addAlbum2}, {&rmvArtistFred, &updArtistSam}, {&addArtistFred}, {&addAlbumFred}}
	//create specific timestamps
	var timestamps = []clocksi.ClockSiTimestamp{
		{VectorClock: map[int16]int64{111: 1, 222: 0}},
		{VectorClock: map[int16]int64{111: 2, 222: 0}},
		{VectorClock: map[int16]int64{111: 1, 222: 1}},
		{VectorClock: map[int16]int64{111: 3, 222: 1}},
		{VectorClock: map[int16]int64{111: 2, 222: 2}},
		{VectorClock: map[int16]int64{111: 3, 222: 2}},
		{VectorClock: map[int16]int64{111: 4, 222: 2}}}

	testReplica(crdtR1, 1, opOrderR1, timestamps)
}

func TestNoOpCrdt4() {
	fmt.Println("\n Test start TestNoOpCrdt4")
	crdtR1 := (&NoOpCrdt{}).Initialize(nil, 111).(*NoOpCrdt)

	var opOrderR1 = [][]Operation{{&addArtistSam, &addArtistFred}, {&addAlbum1, &addAlbum2}, {&rmvArtistFred, &updArtistSam}, {&addAlbumFred}}
	//create specific timestamps
	var timestamps = []clocksi.ClockSiTimestamp{
		{VectorClock: map[int16]int64{111: 1, 222: 0}},
		{VectorClock: map[int16]int64{111: 0, 222: 1}},
		{VectorClock: map[int16]int64{111: 2, 222: 1}},
		{VectorClock: map[int16]int64{111: 1, 222: 2}},
		{VectorClock: map[int16]int64{111: 3, 222: 2}},
		{VectorClock: map[int16]int64{111: 2, 222: 3}},
		{VectorClock: map[int16]int64{111: 4, 222: 4}}}

	testReplica(crdtR1, 1, opOrderR1, timestamps)
}

// Func that carries out basic test structure
func testReplica(c *NoOpCrdt, replicaNum int, opBlocks [][]Operation, timestamps []clocksi.ClockSiTimestamp) {
	fmt.Printf("Starting Replica%v:\n", replicaNum)
	processOps(c, opBlocks, timestamps)
	fmt.Printf("Final result, Replica%v:\n", replicaNum)
	PrintNoOpCrdt(c, fmt.Sprintf("Replica%v", replicaNum))
}

// Func that prepares and builds the graph in the crdt taking into account operation conflicts
func processOps(c *NoOpCrdt, opBlocks [][]Operation, timestamps []clocksi.ClockSiTimestamp) {
	i := 0
	for _, opBlock := range opBlocks {
		conflictMsgs := []Message{}
		j := i
		for _, op := range opBlock { //Prepare messages for all conflicting messages before applying downstream
			conflictMsgs = append(conflictMsgs, c.Update(op).(Message))
			j++
		}
		for k := i; k < j; k++ {
			c.Downstream(timestamps[k], conflictMsgs[k-i])
		}
		i = j
	}
}

// Func that prints out contents of a NoOpCrdt
func PrintNoOpCrdt(c *NoOpCrdt, name string) {
	fmt.Println("--------------------", name)
	fmt.Println("Operations:")
	for _, node := range c.NodeArr {
		fmt.Println(node.Value.Op, node.IsNoOp)
		/*
			fmt.Println("Blocks----")
			for _, op := range node.Value.BlockedOps {
				fmt.Println("--", op)
			}
			fmt.Println("--------")
		*/
	}
	PrintMusicData(c.Read(nil, nil).(*MusicData))
}

// Func that prints out the data (Music Data -> Artists and Albums) for a NoOpCrdt using MusicData
func PrintMusicData(d *MusicData) {
	fmt.Println("State:")
	for artist, val := range *d {
		fmt.Println("Artist:", artist)
		fmt.Println("-- Albums: ", val.Len())
		for _, album := range val.Keys() {
			fmt.Println("----", album)
		}
	}
}
