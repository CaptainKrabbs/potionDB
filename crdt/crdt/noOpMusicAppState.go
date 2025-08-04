package crdt

import (
	"fmt"
	"potionDB/crdt/graphPackages/hashset"
	"potionDB/crdt/proto"
)

var endtag string = "<>"

// App types
type Artist string
type Album string

type NoOpState interface {
	State
	ProtoState
	Copy() NoOpState
	GetStateType() int32

	Serialize() [][]byte
	Deserialize([][]byte) NoOpState
}

type MusicMap map[Artist]*hashset.HashSet[Album]

// State type
type MusicState struct {
	State MusicMap
}

func InitMusicState() *MusicState {
	return &MusicState{State: make(map[Artist]*hashset.HashSet[Album])}
}

func (d *MusicState) GetStateType() int32 { return MusicStateType }

func (d *MusicState) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (d *MusicState) GetREADType() proto.READType { return proto.READType_FULL }

// Returns a deep copy of the artists and albums
// using copy() when copying the slice since the slice elements are value type
func (d *MusicState) Copy() NoOpState {
	newMap := make(MusicMap)
	for artist, artistAlbums := range d.State {

		var newAlbums *hashset.HashSet[Album] = hashset.New[Album]()
		for _, album := range artistAlbums.Keys() {
			newAlbums.Add(album)
		}
		newMap[artist] = newAlbums
	}
	return &MusicState{State: newMap}
}

func (d *MusicState) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	for _, musicOp := range musicOpStatic {
		if protobuf.GetNoop().GetOpCode() == musicOp.GetOpCode(){
			return musicOp.FromUpdateObject(protobuf)
		}
	}
	return nil
}

func (d *MusicState) ToReadResp() (protobuf *proto.ApbReadObjectResp) {
	stateType := d.GetStateType()
	if protobuf.GetNoop().GetStateType() != stateType {
		fmt.Printf("Error occurred: Invalid State Type. Attempting to serialize State of code: %v from protobuf of code: %v\n", d.GetStateType(), protobuf.GetNoop().GetStateType())
		return nil
	}
	return &proto.ApbReadObjectResp{Noop: &proto.ApbGetNoOpResp{StateType: &stateType, StateData: d.Serialize()}}
}

func (d *MusicState) FromReadResp(protobuf *proto.ApbReadObjectResp) State {
	return d.Deserialize(protobuf.GetNoop().GetStateData())
}

//Auxiliary functions

//Converts the map of Artists and Albums to an [][]byte
//OLD FORMAT: Format of Serialization: Number of Artists, name of artist, number of albums for artist, album name1, album name2..., name of artist2, etc...
//NEW FORMAT: Format of Serialization: name of artist, name of album, name of next album (endtag symbolizes end of album sequence)
func (d *MusicState) Serialize() (bytes [][]byte) {
	bytes = [][]byte{}
	for artist, artistAlbums := range d.State {
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
func (d *MusicState) Deserialize(bytes [][]byte) NoOpState {
	newMap := make(MusicMap)
	var i int = 0
	for i < len(bytes) {
		artist := Artist(bytes[i])
		i++
		var newAlbums *hashset.HashSet[Album] = hashset.New[Album]()
		for album := string(bytes[i]); album != endtag; i++ {
			newAlbums.Add(Album(album))
		}
		i++
		newMap[artist] = newAlbums
	}
	return &MusicState{State: newMap}
}

func GetNoOpStateFromProto(protobuf *proto.ProtoNoOpState) NoOpState {
	switch protobuf.GetStateType() {
	case MusicStateType:
		return (&MusicState{}).Deserialize(protobuf.GetStateData())
	default:
		fmt.Printf("[Error][noOpMusicAppState] StateType from ProtoNoOpState not valid for a NoOpCrdt: %v\n", protobuf.GetStateType())
		return nil
	}
}