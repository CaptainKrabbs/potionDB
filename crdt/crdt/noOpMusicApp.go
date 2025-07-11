package crdt

import (
	"fmt"
	"potionDB/crdt/graphPackages/hashset"
	"potionDB/crdt/proto"
)

// App types
type Artist string
type Album string

// State type
type MusicData map[Artist]*hashset.HashSet[Album]

func (d *MusicData) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (d *MusicData) GetREADType() proto.READType { return proto.READType_FULL }

// Returns a deep copy of the artists and albums
// using copy() when copying the slice since the slice elements are value type
func (d *MusicData) Copy() MusicData {
	newMap := make(MusicData)
	for artist, artistAlbums := range *d {

		var newAlbums *hashset.HashSet[Album] = hashset.New[Album]()
		for _, album := range artistAlbums.Keys() {
			newAlbums.Add(album)
		}
		newMap[artist] = newAlbums
	}
	return newMap
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
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		_, artistExists := (*artistAlbums)[a.ArtistName]
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
		(*artistAlbums)[a.ArtistName] = hashset.New[Album]()
	}
}

func (a *AddArtist) String() string {
	return fmt.Sprintf("Operation: AddArtist(%v)", a.ArtistName)
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
		_, artistExists := (*artistAlbums)[a.ArtistName]
		return artistExists
	}
	return false
}

func (a *RmvArtist) BlockGenerator() []Operation {
	return nil
}

func (a *RmvArtist) Process(state State) {
	if artistAlbums, isMusicData := state.(*MusicData); isMusicData {
		delete(*artistAlbums, a.ArtistName)
	}
}

func (a *RmvArtist) String() string {
	return fmt.Sprintf("Operation: RmvArtist(%v)", a.ArtistName)
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
		_, artistExists := (*artistAlbums)[a.ArtistName]
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
		albums, artistExists := (*artistAlbums)[a.ArtistName]
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
		(*artistAlbums)[a.ArtistName].Add(a.AlbumName)
	}
}

func (a *AddAlbum) String() string {
	return fmt.Sprintf("Operation: AddAlbum(%v, %v)", a.AlbumName, a.ArtistName)
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
		albums, artistExists := (*artistAlbums)[a.ArtistName]
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
		(*artistAlbums)[a.ArtistName].Delete(a.AlbumName)
	}
}

func (a *RmvAlbum) String() string {
	return fmt.Sprintf("Operation: RmvAlbum(%v, %v)", a.ArtistName, a.AlbumName)
}

// operations
func (args AddArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args RmvArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args UpdArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (args AddAlbum) GetCRDTType() proto.CRDTType  { return proto.CRDTType_NOOP }
func (args RmvAlbum) GetCRDTType() proto.CRDTType  { return proto.CRDTType_NOOP }
