package crdt

import (
	"fmt"
	"potionDB/crdt/clocksi"
	"testing"
)

func TestNoOpCrdt1(t *testing.T) {
	crdtR1 := (&NoOpCrdt{}).Initialize(nil, 111).(*NoOpCrdt)
	crdtR2 := (&NoOpCrdt{}).Initialize(nil, 222).(*NoOpCrdt)
	//newDownstreamR1 := make([]DownstreamArguments, 0, 5)
	//newDownstreamR2 := make([]DownstreamArguments, 0, 5)

	addArtistSam := AddArtist{ArtistName: "Sam"}
	addAlbum1 := AddAlbum{AlbumName: "A1", ArtistName: "Sam"}
	addAlbum2 := AddAlbum{AlbumName: "A2", ArtistName: "Sam"}
	updArtistSam := UpdArtist{ArtistName: "Sam"}
	rmvArtistSam := RmvArtist{ArtistName: "Sam"}

	var opOrderR1 = []Operation{&addArtistSam, &addAlbum1, &addAlbum2, &updArtistSam, &rmvArtistSam}
	var opOrderR2 = []Operation{&addArtistSam, &addAlbum1, &addAlbum2, &rmvArtistSam, &updArtistSam}
	//create specific timestamps
	var timestamps = []clocksi.ClockSiTimestamp{
		{VectorClock: map[int16]int64{111: 1, 222: 0}},
		{VectorClock: map[int16]int64{111: 2, 222: 0}},
		{VectorClock: map[int16]int64{111: 1, 222: 1}},
		{VectorClock: map[int16]int64{111: 3, 222: 1}},
		{VectorClock: map[int16]int64{111: 2, 222: 2}}}

	for i, op := range opOrderR1 {
		call := crdtR1.Update(op).(Message)
		crdtR2.Downstream(timestamps[i], call)
	}

	for i, op := range opOrderR2 {
		call := crdtR2.Update(op).(Message)
		crdtR2.Downstream(timestamps[i], call)
	}

	stateR1 := crdtR1.Read(StateReadArguments{}, nil).(*MusicData)
	stateR2 := crdtR2.Read(StateReadArguments{}, nil).(*MusicData)

	fmt.Println(stateR1)
	fmt.Println(stateR2)
}
