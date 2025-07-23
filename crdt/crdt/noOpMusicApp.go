package crdt

import (
	"bytes"
	"fmt"
	"encoding/gob"
	"potionDB/crdt/graphPackages/hashset"
	"potionDB/crdt/proto"
)

// App types
type Artist string
type Album string

type NoOpState interface {
	Copy() NoOpState
	GetCRDTType() proto.CRDTType
	GetREADType() proto.READType
	ToReadResp() *proto.ApbReadObjectResp
	FromReadResp(*proto.ApbReadObjectResp) State
}

type MusicMap map[Artist]*hashset.HashSet[Album]

// State type
type MusicState struct {
	State MusicMap
}

func InitMusicState() *MusicState {
	return &MusicState{State: make(map[Artist]*hashset.HashSet[Album])}
}

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
	return &MusicState{newMap}
}

func (d *MusicState) ToReadResp() (protobuf *proto.ApbReadObjectResp) {
	stateType := proto.NoOpStateType_MUSIC_STATE
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(d); err != nil {
		panic(err)
	}
	return &proto.ApbReadObjectResp{Noop: &proto.ApbGetNoOpResp{Type: &stateType, Value: buf.Bytes()}}
}

func (d *MusicState) FromReadResp(protobuf *proto.ApbReadObjectResp) (state State) {
	var buf bytes.Buffer
	var decodedState NoOpState
	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(&decodedState); err != nil {
		panic(err)
	}
	return decodedState
}

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

func (a *AddArtist) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	updType := proto.NoOpStateType_MUSIC_STATE
	musicUpd := proto.ApbNoOpMusicStateUpdate{AddArtistOp: &proto.ApbNoOpMusicStateAddArtist{ArtistName: (*string)(&a.ArtistName)}}
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{Type: &updType, MusicUpd: &musicUpd}}
}

func (a *AddArtist) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	return AddArtist{ArtistName: Artist(*protobuf.Noop.MusicUpd.AddArtistOp.ArtistName)}
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

func (a *RmvArtist) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	updType := proto.NoOpStateType_MUSIC_STATE
	musicUpd := proto.ApbNoOpMusicStateUpdate{RmvArtistOp: &proto.ApbNoOpMusicStateRmvArtist{ArtistName: (*string)(&a.ArtistName)}}
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{Type: &updType, MusicUpd: &musicUpd}}
}

func (a *RmvArtist) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	return RmvArtist{ArtistName: Artist(*protobuf.Noop.MusicUpd.RmvArtistOp.ArtistName)}
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

func (a *UpdArtist) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	updType := proto.NoOpStateType_MUSIC_STATE
	musicUpd := proto.ApbNoOpMusicStateUpdate{UpdArtistOp: &proto.ApbNoOpMusicStateUpdArtist{ArtistName: (*string)(&a.ArtistName)}}
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{Type: &updType, MusicUpd: &musicUpd}}
}

func (a *UpdArtist) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	return UpdArtist{ArtistName: Artist(*protobuf.Noop.MusicUpd.UpdArtistOp.ArtistName)}
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

func (a *AddAlbum) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	updType := proto.NoOpStateType_MUSIC_STATE
	musicUpd := proto.ApbNoOpMusicStateUpdate{AddAlbumOp: &proto.ApbNoOpMusicStateAddAlbum{ArtistName: (*string)(&a.ArtistName), AlbumName: (*string)(&a.AlbumName)}}
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{Type: &updType, MusicUpd: &musicUpd}}
}

func (a *AddAlbum) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	return AddAlbum{ArtistName: Artist(*protobuf.Noop.MusicUpd.AddAlbumOp.ArtistName), AlbumName: Album(*protobuf.Noop.MusicUpd.AddAlbumOp.AlbumName)}
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

func (a *RmvAlbum) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	updType := proto.NoOpStateType_MUSIC_STATE
	musicUpd := proto.ApbNoOpMusicStateUpdate{RmvAlbumOp: &proto.ApbNoOpMusicStateRmvAlbum{ArtistName: (*string)(&a.ArtistName), AlbumName: (*string)(&a.AlbumName)}}
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{Type: &updType, MusicUpd: &musicUpd}}
}

func (a *RmvAlbum) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	return RmvAlbum{ArtistName: Artist(*protobuf.Noop.MusicUpd.RmvAlbumOp.ArtistName), AlbumName: Album(*protobuf.Noop.MusicUpd.RmvAlbumOp.AlbumName)}
}

// operations
func (args AddArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args RmvArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args UpdArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args AddAlbum) GetCRDTType() proto.CRDTType  { return proto.CRDTType_NOOP }
func (args RmvAlbum) GetCRDTType() proto.CRDTType  { return proto.CRDTType_NOOP }
