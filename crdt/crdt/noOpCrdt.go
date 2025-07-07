package crdt

import (
	"potionDB/crdt/clocksi"
	"potionDB/crdt/proto"

	"potionDB/crdt/graphPackages/comparator"
	"potionDB/crdt/graphPackages/hashset"
)

// Node struct for graph
type Node struct {
	Value  Call
	Edges  []int
}

func (node *Node) addEdge(edgeIdx int) {
	node.Edges = append(node.Edges, edgeIdx)
}

// Add this import statement

type Artist string
type Album string
type Albums hashset.HashSet[Album]
type Data map[Artist]Albums

type NoOpCrdt struct {
	CRDTVM
	Comp       comparator.CallComparator
	ArtistAlbums Data
	NodeArr    []Node
	//Adicionar outros campos
}

// Operations
type Operation interface {
	OpEqual(Operation)bool
}

type AddArtist struct {
	ArtistName string
}
func (a AddArtist)OpEqual(o Operation) bool {
	oConv, ok := o.(AddArtist)
	return ok && a.ArtistName == oConv.ArtistName
}

type RmvArtist struct {
	ArtistName string
}
func (a RmvArtist)OpEqual(o Operation) bool {
	oConv, ok := o.(RmvArtist)
	return ok && a.ArtistName == oConv.ArtistName
}

type UpdArtist struct {
	ArtistName string
}
func (a UpdArtist)OpEqual(o Operation) bool {
	oConv, ok := o.(UpdArtist)
	return ok && a.ArtistName == oConv.ArtistName
}

type AddAlbum struct {
	AlbumName, ArtistName string
}
func (a AddAlbum)OpEqual(o Operation) bool {
	oConv, ok := o.(AddAlbum)
	return ok && a.ArtistName == oConv.ArtistName && a.AlbumName == oConv.AlbumName
}

type RmvAlbum struct {
	AlbumName, ArtistName string
}
func (a RmvAlbum)OpEqual(o Operation) bool {
	oConv, ok := o.(RmvAlbum)
	return ok && a.ArtistName == oConv.ArtistName && a.AlbumName == oConv.AlbumName
}

func (a NoOp)OpEqual(o Operation) bool {
	return false
}

// Message struct that stores an operation and those it blocks
type Message struct {
	Op Operation
	BlockedOps []Operation
}

func (msg Message) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP}
func (msg Message) MustReplicate() bool { return false}

//Converts a message to a call struct given a vector Timestamp
func (m Message) ToCall(time clocksi.Timestamp) Call {
	return Call{Op: m.Op, BlockedOps: m.BlockedOps, Time: time}
}

//struct that stores an operation, those it blocks and a vector clock Timestamp
type Call struct {
	Op Operation
	BlockedOps []Operation
	Time clocksi.Timestamp
}

//check if any blocking operation matches the one in the call.
//
//if so early return true.
//otherwise false.
func (actualCall *Call) Blocks(otherCall *Call) bool {
	for i := range(actualCall.BlockedOps) {
		if (actualCall.BlockedOps[i]).OpEqual(otherCall.Op) {
			return true
		}
	}
	return false
}

func (call Call) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP}
func (call Call) MustReplicate() bool { return false}

//operations
func (args AddArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP}
func (args RmvArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP}
func (args UpdArtist) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP}
func (args AddAlbum) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP}
func (args RmvAlbum) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP}

//command
type Command struct {
	opType proto.CRDTType
	Precondition func(Operation) bool
	Process func(Data, []string) error //takes data and string fields
	BlockGenerator func(Operation) []Operation
}

//Nota: no codigo (e.g., no counterCRDT) quando vires "Effect", nao confundas com o Effect nos no-ops.
//Aqui no PotionDB, effect significa qual o "efeito" que uma operação teve no estado do CRDT.
//Este "efeito" é usado pelos CRDTs pela gestão de versões, de modo a poder-se recalcular versões antigas dos objectos.

func (crdt *NoOpCrdt) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }

func (crdt *NoOpCrdt) Initialize(startTs *clocksi.Timestamp, replicaID int16) (newCrdt CRDT) {
	return &NoOpCrdt{
		CRDTVM: (&genericInversibleCRDT{}).initialize(startTs, crdt.undoEffect, crdt.reapplyOp, crdt.notifyRebuiltComplete),
		Comp: comparator.CallComparator{},
		ArtistAlbums: make(Data),
		NodeArr: []Node{},
		//Adicionar outros campos
	}
}

// Used to initialize when building a CRDT from a remote snapshot
func (crdt *NoOpCrdt) initializeFromSnapshot(startTs *clocksi.Timestamp, replicaID int16) (sameCRDT *NoOpCrdt) {
	crdt.CRDTVM = (&genericInversibleCRDT{}).initialize(startTs, crdt.undoEffect, crdt.reapplyOp, crdt.notifyRebuiltComplete)
	return crdt
}

func (crdt *NoOpCrdt) IsBigCRDT() bool { return false }

func (crdt *NoOpCrdt) Read(args ReadArguments, updsNotYetApplied []UpdateArguments) (state State) {
	//TODO. Deixa este método para depois do Update e Downstream. Ignora o updsNotYetApplied e o args.
	//O state deves ser tu proprio a definir, apenas precisa de implementar dois métodos:
	//GetCRDTType() proto.CRDTType:		{return proto.CRDTType_NOOP}
	//GetREADType() proto.READType: 	{return proto.READType_FULL}
	return nil
}

// Prepare
func (crdt *NoOpCrdt) Update(args UpdateArguments) (downstreamArgs DownstreamArguments) {
	//TODO: Este e o prepare. Faz aqui o codigo necessario para gerar os blocks e afins.
	//No final, deves retornar a operacao a ser executada na fase do effect.
	switch typedArgs := args.(type) {
	case AddArtist:
		downstreamArgs = Message{
			Op: typedArgs,
			BlockedOps: nil,
		}
	case RmvArtist:
		downstreamArgs = Message{
			Op: typedArgs,
			BlockedOps: nil,
		}
	case UpdArtist:
		downstreamArgs = Message{
			Op: typedArgs,
			BlockedOps: []Operation{
			RmvArtist{ArtistName: typedArgs.ArtistName},
		}}
	case AddAlbum:
		//check precondition(!artistExists)
		_, artistExists := crdt.ArtistAlbums[Artist(typedArgs.ArtistName)]
		if artistExists {
		downstreamArgs = Message {
			Op: typedArgs,
			BlockedOps: []Operation{
			RmvArtist{ArtistName: typedArgs.ArtistName},
		}}
		} else {
			downstreamArgs = Message{Op: NoOp{}, BlockedOps: nil}
		}
	case RmvAlbum:
		downstreamArgs = Message {
			Op: typedArgs,
			BlockedOps: []Operation{
			RmvArtist{ArtistName: typedArgs.ArtistName},
		}}
	case NoOp:
		downstreamArgs = Message{Op: NoOp{}, BlockedOps: nil}
	}
	return
}

// Effect
func (crdt *NoOpCrdt) Downstream(updTs clocksi.Timestamp, downstreamArgs DownstreamArguments) (otherDownstreamArgs DownstreamArguments) {
	//TODO: Este é o effect. Regra geral aqui costumo chamar um metodo auxiliar "applyDownstream" que aplica mesmo os efeitos da operacao
	//Nesta fase como precisas do updTs, diria que podes fazer aqui a logica do effect (nomeadamente gerar os blocks)
	//A operacao em si pode ser aplicada na applyDownstream
	//No final podes retornar nil.
	//Por agora deixa a linha seguinte comentada: mais tarde será necessária.
	//crdt.addToHistory(&updTs, &downstreamArgs, effect)

	//Expected that the argument here is a Message struct, which needs to be converted into a call
	switch msgType := downstreamArgs.(type) {
	case Message:
		crdt.applyDownstream(msgType.ToCall(updTs)) //Podes alterar esta se precisares
	}
	return nil
}

// Nota: este metodo nao faz parte da interface CRDT, por isso se precisares podes apaga-lo.
func (crdt *NoOpCrdt) applyDownstream(downstreamArgs DownstreamArguments) (effect *Effect) {
	//TODO: Metodo auxiliar do downstream.
	//Por agora em termos de retorno podes deixar o que pus aqui
	newCall := downstreamArgs.(Call)
	newNodeIdx := len(crdt.NodeArr)
	// for v ∈ V
	for i := range crdt.NodeArr {
		compVal, err := crdt.Comp.Compare(&crdt.NodeArr[i].Value, &newCall)
		// if v ≺ c then
		if compVal < 0 { //(compVal != 0 means err is nil)
			// E <- E ∪ {<v,c>}
			(&crdt.NodeArr[i]).addEdge(newNodeIdx)
			// else if v ∥ c ∧ opsConflict(v, c) then (v and c are concurrent and callTypeoperation is among blocked operations)
		} else if compVal == 0 && err == nil {
			if crdt.NodeArr[i].Value.Blocks(&newCall) { //new Call is a No Op
				//mark call as no-op
				newCall.Op = NoOp{}
			}
			if newCall.Blocks(&crdt.NodeArr[i].Value) { //Existing Call is a No Op
				crdt.NodeArr[i].Value.Op = NoOp{}
			}
		}
		//V ← V ∪ {c}
		crdt.NodeArr = append(crdt.NodeArr, Node{Value: newCall, Edges: nil})
	}

	var effectV Effect = NoEffect{}
	return &effectV
}

func (crdt *NoOpCrdt) IsOperationWellTyped(args UpdateArguments) (ok bool, err error) {
	return true, nil
}

func (crdt *NoOpCrdt) Copy() (copyCRDT InversibleCRDT) {
	newCRDT := NoOpCrdt{
		CRDTVM: crdt.CRDTVM.copy(),
		//Adicionar outros campos que pertençam ao NoOpCrdt
	}
	return &newCRDT
}

func (crdt *NoOpCrdt) RebuildCRDTToVersion(targetTs clocksi.Timestamp) {
	//TODO: Might be worth it to make one specific for registers
	crdt.CRDTVM.rebuildCRDTToVersion(targetTs)
}

func (crdt *NoOpCrdt) reapplyOp(updArgs DownstreamArguments) (effect *Effect) {
	return crdt.applyDownstream(updArgs)
}

func (crdt *NoOpCrdt) undoEffect(effect *Effect) {
	//Do not fill this function for now.
}

func (crdt *NoOpCrdt) notifyRebuiltComplete(currTs *clocksi.Timestamp) {}

func (crdt *NoOpCrdt) GetCRDT() CRDT { return crdt }
