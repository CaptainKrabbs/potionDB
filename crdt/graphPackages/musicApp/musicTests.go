package musicApp
/*
import (
	"errors"
	"fmt"
)

const (
	addArtist = "addArtist"
	addArtistNumFields = 1

	rmvArtist = "rmvArtist"
	rmvArtistNumFields = 1

	updArtist = "updArtist"
	updArtistNumFields = 1

	addAlbum  = "addAlbum"
	addAlbumNumFields = 2

	rmvAlbum  = "rmvAlbum"
	rmvAlbumNumFields = 2
)

//Precondition funcs

//Placeholder default: no preconditions
func noPrecondition(data app.Data, op operation.Operation) (bool, error) {
	return true, nil
}

//Returns a function that checks if an artist for a name at a given field index in an operation exists
func artistExists(idx int) func (app.Data, operation.Operation) (bool, error) {
	return func (artistAlbums app.Data, op operation.Operation) (bool, error) {
		if idx >= len(op.Fields) {
			return false, errors.New("precondition error: Index out of range for number of fields in operation")
		}
		_, ok := artistAlbums[op.Fields[idx]]
		return ok, nil
	}
}

//Process funcs
//Placeholder default: no changes made to app data
func noEffect(data app.Data, fields []string) error {
	return nil
}

//aux method to verify number of fields
func fieldNumCheck(expected int, fields []string, opName string) error {
	if len(fields) != expected { //technically already checked in pre-conditions...
		return fmt.Errorf("process error(%v): invalid number of fields", opName)
	}
	return nil
}

func addArtistFn(data app.Data, fields []string) error {
	if err := fieldNumCheck(addArtistNumFields, fields, addArtist); err != nil { return err } //technically checked in precondition

	artistName := fields[0]
	if _, hasArtist := data[fields[0]]; hasArtist {
		return errors.New("process error(addArtist): artist already exists")
	}
	data[artistName] = hashset.New[string]()
	return nil
}

func rmvArtistFn(data app.Data, fields []string) error {
	if err := fieldNumCheck(rmvArtistNumFields, fields, rmvArtist); err != nil { return err }

	artistName := fields[0]
	delete (data, artistName)
	return nil
}

func addAlbumFn(data app.Data, fields []string) error {
	if err := fieldNumCheck(addAlbumNumFields, fields, addAlbum); err != nil { return err } //technically checked in precondition

	artistName, albumName := fields[1], fields[0]
	if _, hasArtist := data[artistName]; !hasArtist {
		return errors.New("process error(addAlbum): artist doesn't exist")
	} else	if data[artistName].Contains(albumName) {
		return errors.New("process error(addAlbum): album already exists")
	}
	data[artistName].Add(albumName)
	return nil
}

func rmvAlbumFn(data app.Data, fields []string) error {
	if err := fieldNumCheck(addAlbumNumFields, fields, addAlbum); err != nil { return err } //technically checked in precondition
	//fields = {albumName, artistName}
	artistName := fields[1]
	if _, hasArtist := data[artistName]; !hasArtist {
		return nil //no artist therefore no album, not an error
	}
	data[artistName].Delete(fields[0])
	return nil
}

//BlockGenerator funcs
func noBlock(o operation.Operation) []operation.Operation {
	return nil
}

//Block RmvArtist using Field at index 0 and 1 respectively
func blockRmvArtist0(o operation.Operation) []operation.Operation {
	return []operation.Operation{
			{Name: rmvArtist, Fields: []string{o.Fields[0]}},
		}
}

func blockRmvArtist1(o operation.Operation) []operation.Operation {
	return []operation.Operation{
			{Name: rmvArtist, Fields: []string{o.Fields[1]}},
		}
}

//Return Music App with example commands
func CreateMusicApp() app.MusicApp {
	musicApp := app.InitializeMusicApp()
	musicApp.AddCommand(addArtist, addArtistNumFields, noPrecondition, addArtistFn, noBlock)
	musicApp.AddCommand(rmvArtist, addArtistNumFields, noPrecondition, rmvArtistFn, noBlock)
	musicApp.AddCommand(updArtist, addArtistNumFields, noPrecondition, noEffect, blockRmvArtist0)
	musicApp.AddCommand(addAlbum, addAlbumNumFields, artistExists(1), addAlbumFn, blockRmvArtist1)
	musicApp.AddCommand(rmvAlbum, rmvAlbumNumFields, noPrecondition, rmvAlbumFn, blockRmvArtist1)

	err := musicApp.AddCommand(rmvAlbum, rmvAlbumNumFields, noPrecondition, noEffect, blockRmvArtist0)
	fmt.Println(err)
	err = musicApp.AddCommand(rmvAlbum, -1, noPrecondition, noEffect, blockRmvArtist0)
	fmt.Println(err)
	return musicApp
}

func PrepReplica1CallOrder(calls *[]operation.Call) {
	*calls = append(*calls,
		operation.Call{
			Op:   operation.Operation{Name: addArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{1, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A1", "Sam"}},
			Time: operation.Timestamp{VectorClock: []int{2, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A2", "Sam"}},
			Time: operation.Timestamp{VectorClock: []int{1, 1}}},
		operation.Call{
			Op:   operation.Operation{Name: updArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{3, 1}}},
		operation.Call{
			Op:   operation.Operation{Name: rmvArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{2, 2}}},
	)
}

func PrepReplica2CallOrder(ops *[]operation.Call) {
	*ops = append(*ops,
		operation.Call{
			Op:   operation.Operation{Name: addArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{1, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A1", "Sam"}},
			Time: operation.Timestamp{VectorClock: []int{2, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A2", "Sam"}},
			Time: operation.Timestamp{VectorClock: []int{1, 1}}},
		operation.Call{
			Op:   operation.Operation{Name: rmvArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{2, 2}}},
		operation.Call{
			Op:   operation.Operation{Name: updArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{3, 1}}},
	)
}

func PrepWrongNumFields(ops *[]operation.Call) {
	*ops = append(*ops,
		operation.Call{
			Op:   operation.Operation{Name: addArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{1, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A1", "Sam", "goobledy"}},
			Time: operation.Timestamp{VectorClock: []int{2, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A2", "Sam"}},
			Time: operation.Timestamp{VectorClock: []int{1, 1}}},
		operation.Call{
			Op:   operation.Operation{Name: rmvArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{2, 2}}},
		operation.Call{
			Op:   operation.Operation{Name: updArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{3, 1}}},
	)
}

func PrepWrongNameComm(ops *[]operation.Call) {
	*ops = append(*ops,
		operation.Call{
			Op:   operation.Operation{Name: addArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{1, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A1", "Sam"}},
			Time: operation.Timestamp{VectorClock: []int{2, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A2", "Sam"}},
			Time: operation.Timestamp{VectorClock: []int{1, 1}}},
		operation.Call{
			Op:   operation.Operation{Name: "boostArtist", Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{2, 2}}},
		operation.Call{
			Op:   operation.Operation{Name: updArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{3, 1}}},
	)
}

func PrepFailPrecondition(ops *[]operation.Call) {
	*ops = append(*ops,
		operation.Call{
			Op:   operation.Operation{Name: addArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{1, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A1", "Sam"}},
			Time: operation.Timestamp{VectorClock: []int{2, 0}}},
		operation.Call{
			Op:   operation.Operation{Name: addAlbum, Fields: []string{"A2", "Josh"}},
			Time: operation.Timestamp{VectorClock: []int{1, 1}}},
		operation.Call{
			Op:   operation.Operation{Name: "boostArtist", Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{2, 2}}},
		operation.Call{
			Op:   operation.Operation{Name: updArtist, Fields: []string{"Sam"}},
			Time: operation.Timestamp{VectorClock: []int{3, 1}}},
	)
}

//Tests the GetState() for a given slice of Calls using the block table defined above
func TestState(g *graph.OpGraph, calls []operation.Call) {
	fmt.Println("New graph:")
	fmt.Println(g)
	fmt.Println(g.IsEmpty())

	for i := range( calls ) {
		var callP *operation.Call = &calls[i]
		msg, err := g.Prepare(callP.Op)
		if err != nil {
			fmt.Println("Error:", err)
			return
		} else {
			fmt.Println("Message", i, msg)
			//Call with blocks
			var call operation.Call = msg.ToCall(callP.Time)
			g.Effect(call)
		}
		
	}
	fmt.Println(g)
	fmt.Println("State pre ops")
	g.App.ArtistAlbums.Print()

	state, err := g.GetState()
	if err == nil {
		fmt.Println("State post ops:")
		state.Print()
	} else {
		fmt.Println("Error:", err)
	}
}

*/