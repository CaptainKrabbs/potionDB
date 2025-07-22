package crdt

import (
	"fmt"
	"potionDB/crdt/graphPackages/hashset"
	"potionDB/crdt/proto"
)

// App types
type Artist string
type Album string

type CrdtData interface {
	Copy() CrdtData
	GetCRDTType() proto.CRDTType
	GetREADType() proto.READType

	//for protobuf
	ProcessFromUpdateObject(*proto.ApbUpdateOperation) UpdateArguments
}

type MusicMap map[Artist]*hashset.HashSet[Album]

// State type
type MusicData struct {
	Data MusicMap
}

func InitMusicData() *MusicData {
	return &MusicData{Data: make(map[Artist]*hashset.HashSet[Album])}
}

func (d *MusicData) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (d *MusicData) GetREADType() proto.READType { return proto.READType_FULL }

// Returns a deep copy of the artists and albums
// using copy() when copying the slice since the slice elements are value type
func (d *MusicData) Copy() CrdtData {
	newMap := make(MusicMap)
	for artist, artistAlbums := range d.Data {

		var newAlbums *hashset.HashSet[Album] = hashset.New[Album]()
		for _, album := range artistAlbums.Keys() {
			newAlbums.Add(album)
		}
		newMap[artist] = newAlbums
	}
	return &MusicData{newMap}
}


func (d *MusicData) ProcessFromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if opCode := protobuf.NoOp.GetOpCode(); opCode >= 0 || opCode < len(musicOpsStatic) {
		return musicOpsStatic[opCode].FromUpdateObject(protobuf)
	} else {
		//Would technically be an error just doing this as a placeholder.
		return (&NoOp{}).FromUpdateObject(protobuf)
	}
}

//Operation Codes
var (
	musicOpsStatic []Operation = []Operation{&NoOp{}, &AddArtist{}, &RmvArtist{}, &UpdArtist{}, &AddAlbum{}, &RmvAlbum{}}
	AddArtistOpCode int32 = 1
	RmvArtistOpCode int32 = 2
	UpdArtistOpCode int32 = 3
	AddAlbumOpCode int32 = 4
	RmvAlbumOpCode int32 = 5

	addArtistNumParams int32 = 1
	rmvArtistNumParams int32 = 1
	updArtistNumParams int32 = 1
	addAlbumNumParams int32 = 2
	rmvAlbumNumParams int32 = 2

	artistNameParamIdx int = 0
	albumNameParamIdx int = 1
)

// Operation implementation structs
// BlockGeneration on Delete.Loses approach principle on conflict

// --------------------AddArtist
type AddArtist struct {
	ArtistName Artist
}

func (a *AddArtist) OpEqual(o Operation) bool {
	oConv, ok := o.(*AddArtist)
	return ok && a.ArtistName == oConv.ArtistName
}

func (a *AddArtist) Copy() Operation {
	return &AddArtist{ArtistName: a.ArtistName}
}

func (a *AddArtist) Precondition(state State) bool {
	// Precondition: !artistExists
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		_, artistExists := artistAlbums.Data[a.ArtistName]
		return !artistExists
	}
	// returns false if the state isn't a type used by musicApp (shouldn't be triggered)
	return false
}

func (a *AddArtist) BlockGenerator() []Operation {
	return nil
}

func (a *AddArtist) Process(state State) {
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		artistAlbums.Data[a.ArtistName] = hashset.New[Album]()
	}
}

func (a *AddArtist) String() string {
	return fmt.Sprintf("Operation: AddArtist(%v)", a.ArtistName)
}

func (a *AddArtist) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{op_code: AddArtistOpCode, parameters: []byte(a.ArtistName)}}
}

func (a *AddArtist) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if len(protobuf.NoOp.GetParams()) != addArtistNumParams {
		return NoOp{}
	}
	return AddArtist{ArtistName: protobuf.NoOp.GetParams()[artistNameParamIdx].(string)}
}

// --------------------RmvArtist
type RmvArtist struct {
	ArtistName Artist
}

func (a *RmvArtist) OpEqual(o Operation) bool {
	oConv, ok := o.(*RmvArtist)
	return ok && a.ArtistName == oConv.ArtistName
}

func (a *RmvArtist) Copy() Operation {
	return &RmvArtist{ArtistName: a.ArtistName}
}

func (a *RmvArtist) Precondition(state State) bool {
	// Precondition: artistExists
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		_, artistExists := artistAlbums.Data[a.ArtistName]
		return artistExists
	}
	return false
}

func (a *RmvArtist) BlockGenerator() []Operation {
	return nil
}

func (a *RmvArtist) Process(state State) {
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		delete(artistAlbums.Data, a.ArtistName)
	}
}

func (a *RmvArtist) String() string {
	return fmt.Sprintf("Operation: RmvArtist(%v)", a.ArtistName)
}

func (a *RmvArtist) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{op_code: RmvArtistOpCode, parameters: [][]byte{[]byte(a.ArtistName)}}}
}

func (a *RmvArtist) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if len(protobuf.NoOp.GetParams()) != rmvArtistNumParams {
		return NoOp{}
	}
	return RmvArtist{ArtistName: protobuf.NoOp.GetParams()[artistNameParamIdx].(string)}
}

// --------------------UpdArtist
type UpdArtist struct {
	ArtistName Artist
}

func (a *UpdArtist) OpEqual(o Operation) bool {
	oConv, ok := o.(*UpdArtist)
	return ok && a.ArtistName == oConv.ArtistName
}

func (a *UpdArtist) Copy() Operation {
	return &UpdArtist{ArtistName: a.ArtistName}
}

func (a *UpdArtist) Precondition(state State) bool {
	// Precondition: artistExists
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		_, artistExists := artistAlbums.Data[a.ArtistName]
		return artistExists
	}
	return false
}

func (a *UpdArtist) BlockGenerator() []Operation {
	return []Operation{
		&RmvArtist{ArtistName: a.ArtistName},
	}
}

func (a *UpdArtist) Process(state State) {
}

func (a *UpdArtist) String() string {
	return fmt.Sprintf("Operation: UpdArtist(%v)", a.ArtistName)
}

func (a *UpdArtist) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{op_code: RmvArtistOpCode, parameters: [][]byte{[]byte(a.ArtistName)}}}
}

func (a *UpdArtist) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if len(protobuf.NoOp.GetParams()) != updArtistNumParams {
		return NoOp{}
	}
	return UpdArtist{ArtistName: protobuf.NoOp.GetParams()[artistNameParamIdx].(string)}
}

// --------------------AddAlbum
type AddAlbum struct {
	AlbumName  Album
	ArtistName Artist
}

func (a *AddAlbum) OpEqual(o Operation) bool {
	oConv, ok := o.(*AddAlbum)
	return ok && a.ArtistName == oConv.ArtistName && a.AlbumName == oConv.AlbumName
}

func (a *AddAlbum) Copy() Operation {
	return &AddAlbum{ArtistName: a.ArtistName, AlbumName: a.AlbumName}
}

func (a *AddAlbum) Precondition(state State) bool {
	// Precondition: artistExists && !albumExists
	/*
		The actual precondition would just be albumExists.
		However, since a map of Artist -> Artist's albums
		is being used for the State.
		To check an album exists we need to get the albums
		from the value at the artist name key.
	*/
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		albums, artistExists := artistAlbums.Data[a.ArtistName]
		return artistExists && !albums.Contains(a.AlbumName)
	}
	return false
}

func (a *AddAlbum) BlockGenerator() []Operation {
	return []Operation{
		&RmvArtist{ArtistName: a.ArtistName},
	}
}

func (a *AddAlbum) Process(state State) {
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		artistAlbums.Data[a.ArtistName].Add(a.AlbumName)
	}
}

func (a *AddAlbum) String() string {
	return fmt.Sprintf("Operation: AddAlbum(%v, %v)", a.AlbumName, a.ArtistName)
}

func (a *AddAlbum) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{op_code: RmvArtistOpCode, parameters: [][]byte{[]byte(a.ArtistName)}}}
}

func (a *AddAlbum) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if len(protobuf.NoOp.GetParams()) != addAlbumNumParams {
		return NoOp{}
	}
	return AddAlbum{ArtistName: protobuf.NoOp.GetParams()[artistNameParamIdx].(string), AlbumName: protobuf.NoOp.GetParams()[albumNameParamIdx].(string)}
}

// --------------------RmvAlbum
type RmvAlbum struct {
	AlbumName  Album
	ArtistName Artist
}

func (a *RmvAlbum) OpEqual(o Operation) bool {
	oConv, ok := o.(*RmvAlbum)
	return ok && a.ArtistName == oConv.ArtistName && a.AlbumName == oConv.AlbumName
}

func (a *RmvAlbum) Copy() Operation {
	return &RmvAlbum{ArtistName: a.ArtistName, AlbumName: a.AlbumName}
}

func (a *RmvAlbum) Precondition(state State) bool {
	// Precondition: artistExists && albumExists
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		albums, artistExists := artistAlbums.Data[a.ArtistName]
		return artistExists && albums.Contains(a.AlbumName)
	}
	return false
}

func (a *RmvAlbum) BlockGenerator() []Operation {
	return []Operation{
		&RmvArtist{ArtistName: a.ArtistName},
	}
}

func (a *RmvAlbum) Process(state State) {
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		artistAlbums.Data[a.ArtistName].Delete(a.AlbumName)
	}
}

func (a *RmvAlbum) String() string {
	return fmt.Sprintf("Operation: RmvAlbum(%v, %v)", a.ArtistName, a.AlbumName)
}

func (a *RmvAlbum) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{op_code: RmvArtistOpCode, parameters: [][]byte{[]byte(a.ArtistName)}}}
}

func (a *RmvAlbum) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if len(protobuf.NoOp.GetParams()) != rmvAlbumNumParams {
		return NoOp{}
	}
	return RmvAlbum{ArtistName: protobuf.NoOp.GetParams()[artistNameParamIdx].(string), AlbumName: protobuf.NoOp.GetParams()[albumNameParamIdx].(string)}
}

// operations
func (args AddArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args RmvArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args UpdArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args AddAlbum) GetCRDTType() proto.CRDTType  { return proto.CRDTType_NOOP }
func (args RmvAlbum) GetCRDTType() proto.CRDTType  { return proto.CRDTType_NOOP }
