package crdt

import (
	"fmt"
	"potionDB/crdt/graphPackages/hashset"
	"potionDB/crdt/proto"
)

var (
	MusicStateType int32 = 1
	//Operation codes for App using MusicState as NoOpState
	AddArtistOpCode int32 = 1
	RmvArtistOpCode int32 = 2
	UpdArtistOpCode int32 = 3
	AddAlbumOpCode int32 = 4
	RmvAlbumOpCode int32 = 5

	//Required number of params per operation
	AddArtistNumParams = 1
	RmvArtistNumParams = 1
	UpdArtistNumParams = 1
	AddAlbumNumParams = 2
	RmvAlbumNumParams = 2

	//Expected Parameter index
	ArtistNameParamIdx = 0
	AlbumNameParamIdx = 1

	musicOpStatic = []Operation{&AddArtist{}, &RmvArtist{}, &UpdArtist{}, &AddAlbum{}, &RmvAlbum{}}
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
	if artistAlbums, isMusicState := state.(*MusicState); isMusicState {
		_, artistExists := artistAlbums.State[a.ArtistName]
		return !artistExists
	}
	// returns false if the state isn't a type used by musicApp (shouldn't be triggered)
	return false
}

func (a *AddArtist) BlockGenerator() []Operation {
	return nil
}

func (a *AddArtist) Process(state State) {
	if artistAlbums, isMusicState := state.(*MusicState); isMusicState {
		artistAlbums.State[a.ArtistName] = hashset.New[Album]()
	}
}

func (a *AddArtist) String() string {
	return fmt.Sprintf("Operation: AddArtist(%v)", a.ArtistName)
}

func (a *AddArtist) GetOpName() string {
	return "AddArtist"
}

func (a *AddArtist) GetStateType() (num int32) {
	return MusicStateType
}

func (a *AddArtist) GetOpCode() (num int32) {
	return AddArtistOpCode
}

func (a *AddArtist) GetNumParams() (num int) {
	return AddArtistNumParams
}

func (a *AddArtist) GetSerializedParams() ([][]byte) {
	return [][]byte{[]byte(a.ArtistName)}
}


func (a *AddArtist) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return ToUpdateObjectFrame(a)
}

func (a *AddArtist) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if !ValidateOpProtobuf(a, protobuf) {return NoOp{}}
	return &AddArtist{ArtistName: Artist(protobuf.Noop.Params[ArtistNameParamIdx])}
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
	if artistAlbums, isMusicState := state.(*MusicState); isMusicState {
		_, artistExists := artistAlbums.State[a.ArtistName]
		return artistExists
	}
	return false
}

func (a *RmvArtist) BlockGenerator() []Operation {
	return nil
}

func (a *RmvArtist) Process(state State) {
	if artistAlbums, isMusicState := state.(*MusicState); isMusicState {
		delete(artistAlbums.State, a.ArtistName)
	}
}

func (a *RmvArtist) String() string {
	return fmt.Sprintf("Operation: RmvArtist(%v)", a.ArtistName)
}

func (a *RmvArtist) GetOpName() string {
	return "RmvArtist"
}

func (a *RmvArtist) GetStateType() (num int32) {
	return MusicStateType
}

func (a *RmvArtist) GetOpCode() (num int32) {
	return RmvArtistOpCode
}

func (a *RmvArtist) GetNumParams() (num int) {
	return RmvArtistNumParams
}

func (a *RmvArtist) GetSerializedParams() ([][]byte) {
	return [][]byte{[]byte(a.ArtistName)}
}


func (a *RmvArtist) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return ToUpdateObjectFrame(a)
}

func (a *RmvArtist) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if !ValidateOpProtobuf(a, protobuf) {return NoOp{}}
	return &RmvArtist{ArtistName: Artist(protobuf.Noop.Params[ArtistNameParamIdx])}
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
	if artistAlbums, isMusicState := state.(*MusicState); isMusicState {
		_, artistExists := artistAlbums.State[a.ArtistName]
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

func (a *UpdArtist) GetOpName() string {
	return "UpdArtist"
}

func (a *UpdArtist) GetStateType() (num int32) {
	return MusicStateType
}

func (a *UpdArtist) GetOpCode() (num int32) {
	return UpdArtistOpCode
}

func (a *UpdArtist) GetNumParams() (num int) {
	return UpdArtistNumParams
}

func (a *UpdArtist) GetSerializedParams() ([][]byte) {
	return [][]byte{[]byte(a.ArtistName)}
}


func (a *UpdArtist) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return ToUpdateObjectFrame(a)
}

func (a *UpdArtist) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if !ValidateOpProtobuf(a, protobuf) {return NoOp{}}
	return &UpdArtist{ArtistName: Artist(protobuf.Noop.Params[ArtistNameParamIdx])}
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
	if artistAlbums, isMusicState := state.(*MusicState); isMusicState {
		albums, artistExists := artistAlbums.State[a.ArtistName]
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
	if artistAlbums, isMusicState := state.(*MusicState); isMusicState {
		artistAlbums.State[a.ArtistName].Add(a.AlbumName)
	}
}

func (a *AddAlbum) String() string {
	return fmt.Sprintf("Operation: AddAlbum(%v, %v)", a.AlbumName, a.ArtistName)
}

func (a *AddAlbum) GetOpName() string {
	return "AddAlbum"
}

func (a *AddAlbum) GetStateType() (num int32) {
	return MusicStateType
}

func (a *AddAlbum) GetOpCode() (num int32) {
	return AddAlbumOpCode
}

func (a *AddAlbum) GetNumParams() (num int) {
	return AddAlbumNumParams
}

func (a *AddAlbum) GetSerializedParams() ([][]byte) {
	return [][]byte{[]byte(a.ArtistName), []byte(a.AlbumName)}
}

func (a *AddAlbum) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return ToUpdateObjectFrame(a)
}

func (a *AddAlbum) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if !ValidateOpProtobuf(a, protobuf) {return NoOp{}}
	return &AddAlbum{ArtistName: Artist(protobuf.Noop.Params[ArtistNameParamIdx]), AlbumName: Album(protobuf.Noop.Params[AlbumNameParamIdx])}
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
	if artistAlbums, isMusicState := state.(*MusicState); isMusicState {
		albums, artistExists := artistAlbums.State[a.ArtistName]
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
	if artistAlbums, isMusicState := state.(*MusicState); isMusicState {
		artistAlbums.State[a.ArtistName].Delete(a.AlbumName)
	}
}

func (a *RmvAlbum) String() string {
	return fmt.Sprintf("Operation: RmvAlbum(%v, %v)", a.ArtistName, a.AlbumName)
}

func (a *RmvAlbum) GetOpName() string {
	return "RmvAlbum"
}

func (a *RmvAlbum) GetStateType() (num int32) {
	return MusicStateType
}

func (a *RmvAlbum) GetOpCode() (num int32) {
	return RmvAlbumOpCode
}

func (a *RmvAlbum) GetNumParams() (num int) {
	return RmvAlbumNumParams
}

func (a *RmvAlbum) GetSerializedParams() ([][]byte) {
	return [][]byte{[]byte(a.ArtistName), []byte(a.AlbumName)}
}

func (a *RmvAlbum) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return ToUpdateObjectFrame(a)
}

func (a *RmvAlbum) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if !ValidateOpProtobuf(a, protobuf) {return NoOp{}}
	return &RmvAlbum{ArtistName: Artist(protobuf.Noop.Params[ArtistNameParamIdx]), AlbumName: Album(protobuf.Noop.Params[AlbumNameParamIdx])}
}

// operations
func (args *AddArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args *RmvArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args *UpdArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args *AddAlbum) GetCRDTType() proto.CRDTType  { return proto.CRDTType_NOOP }
func (args *RmvAlbum) GetCRDTType() proto.CRDTType  { return proto.CRDTType_NOOP }
