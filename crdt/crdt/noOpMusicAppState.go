package crdt

import (
	"potionDB/crdt/graphPackages/hashset"
	"potionDB/crdt/proto"
)

var (
	MusicStateCode int32 = 1
	endtag string = "<>"
)

// App types
type Artist string
type Album string

type MusicMap map[Artist]*hashset.HashSet[Album]

// State type
type MusicState struct {
	State MusicMap
}

/*
Initializes the state data.
Probably unnecessary since with the new DecisionState, the operation that changes
the type of state also initialises it.
*/

func (s *MusicState) Initialize(){
	s.State = make(map[Artist]*hashset.HashSet[Album])
}

func (s *MusicState) GetStateCode() int32 { return MusicStateCode }

func (s *MusicState) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (s *MusicState) GetREADType() proto.READType { return proto.READType_FULL }

// Returns a deep copy of the artists and albums
// using copy() when copying the slice since the slice elements are value type
func (s *MusicState) Copy() NoOpState {
	newMap := make(MusicMap)
	for artist, artistAlbums := range s.State {

		var newAlbums *hashset.HashSet[Album] = hashset.New[Album]()
		for _, album := range artistAlbums.Keys() {
			newAlbums.Add(album)
		}
		newMap[artist] = newAlbums
	}
	return &MusicState{State: newMap}
}

func (s *MusicState) FromUpdateObject(protobuf *proto.ApbUpdateOperation) UpdateArguments {
	if protobuf.GetNoop().GetStateCode() == s.GetStateCode() {
		for _, musicOp := range musicOpStatic {
			if protobuf.GetNoop().GetOpCode() == musicOp.GetOpCode(){
				return musicOp.FromUpdateObject(protobuf)
			}
		}
	}
	return nil
}

func (s *MusicState) ToReadResp() (protobuf *proto.ApbReadObjectResp) {
	return ToReadRespFrame(s)
}

func (s *MusicState) FromReadResp(protobuf *proto.ApbReadObjectResp) State {
	return s.Deserialize(protobuf.GetNoop().GetStateData())
}

//Auxiliary functions

//Converts the map of Artists and Albums to an [][]byte
//OLD FORMAT: Format of Serialization: Number of Artists, name of artist, number of albums for artist, album name1, album name2..., name of artist2, etc...
//NEW FORMAT: Format of Serialization: name of artist, name of album, name of next album (endtag symbolizes end of album sequence)
func (s *MusicState) Serialize() (bytes [][]byte) {
	bytes = [][]byte{}
	for artist, artistAlbums := range s.State {
		//Serialize name of artist and number of albums
		bytes = append(bytes, []byte(artist))
		//Serialize album names
		for _, album := range artistAlbums.Keys() {
			bytes = append(bytes, []byte(album))
		}
		bytes = append(bytes, []byte(endtag))
	}
	return bytes
}

//Deserialization into the MusicState corresponding to the byte matrix.
//Assumes correct serialization.
func (s *MusicState) Deserialize(bytes [][]byte) NoOpState {
	newMap := make(MusicMap)
	var i int = 0
	for i < len(bytes) {
		artist := Artist(bytes[i])
		i++
		var newAlbums *hashset.HashSet[Album] = hashset.New[Album]()
		for album := string(bytes[i]); album != endtag; album = string(bytes[i]) {
			newAlbums.Add(Album(album))
			i++
		}
		i++
		newMap[artist] = newAlbums
	}
	return &MusicState{State: newMap}
}