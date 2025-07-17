package crdt

import (
	"potionDB/crdt/clocksi"
	"potionDB/crdt/proto"
)

// Methods for NoOp, the rest of the operations are in noOpMusicApp.go
func (a *NoOp) OpEqual(o Operation) bool {
	_, ok := o.(*NoOp)
	return ok
}

func (a *NoOp) Copy() Operation {
	return &NoOp{}
}

func (a *NoOp) Precondition(state State) bool {
	return true
}

func (a *NoOp) BlockGenerator() []Operation {
	return nil
}

func (a *NoOp) Process(state State) {
}

func (a *NoOp) String() string {
	return "Operation: NoOp"
}

// Node struct for graph
type Node struct {
	Value  Call
	Edges  []int
	IsNoOp bool
}

func (node *Node) Copy() Node {
	newNode := Node{Value: node.Value.Copy(), Edges: make([]int, len(node.Edges))}
	copy(newNode.Edges, node.Edges) //shallow copy should be enough for ints
	return newNode
}

func (node *Node) addEdge(edgeIdx int) {
	node.Edges = append(node.Edges, edgeIdx)
}

type NoOpCrdt struct {
	CRDTVM
	DataContent CrdtData
	NodeArr      []Node
}

//Nota: no codigo (e.g., no counterCRDT) quando vires "Effect", nao confundas com o Effect nos no-ops.
//Aqui no PotionDB, effect significa qual o "efeito" que uma operação teve no estado do CRDT.
//Este "efeito" é usado pelos CRDTs pela gestão de versões, de modo a poder-se recalcular versões antigas dos objectos.

func (crdt *NoOpCrdt) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }

func (crdt *NoOpCrdt) Initialize(startTs *clocksi.Timestamp, replicaID int16) (newCrdt CRDT) {
	return &NoOpCrdt{
		CRDTVM:       (&genericInversibleCRDT{}).initialize(startTs, crdt.undoEffect, crdt.reapplyOp, crdt.notifyRebuiltComplete),
		DataContent: nil,
		NodeArr:      []Node{},
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

	crdtCpy := crdt.Copy().(*NoOpCrdt)
	stateCpy := crdtCpy.DataContent
	//perform operations on app db
	for _, node := range crdt.NodeArr {
		if !node.IsNoOp {
			node.Value.Op.Process(stateCpy)
		}
	}
	return stateCpy
}

// Prepare
func (crdt *NoOpCrdt) Update(args UpdateArguments) (downstreamArgs DownstreamArguments) {
	//TODO: Este e o prepare. Faz aqui o codigo necessario para gerar os blocks e afins.
	//No final, deves retornar a operacao a ser executada na fase do effect.
	op := args.(Operation)
	//Check precondition for a copy of the crdt.
	if op.Precondition(crdt.Read(nil, nil)) { //TODO: potentially modify this if ReadArguments become relevant
		downstreamArgs = Message{Op: op, BlockedOps: op.BlockGenerator()}
	} else {
		downstreamArgs = Message{Op: &NoOp{}, BlockedOps: nil}
	}
	//Message with NoOp used to detect operations that don't meet preconditions that shouldn't be added to the graph.
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
	msg := downstreamArgs.(Message)
	//Check that operation isn't NoOp (precondition was validated)
	if _, ok := msg.Op.(*NoOp); !ok {
		crdt.applyDownstream(msg.ToCall(updTs)) //Podes alterar esta se precisares
	}
	return nil
}

// Nota: este metodo nao faz parte da interface CRDT, por isso se precisares podes apaga-lo.
func (crdt *NoOpCrdt) applyDownstream(downstreamArgs DownstreamArguments) (effect *Effect) {
	//TODO: Metodo auxiliar do downstream.
	//Por agora em termos de retorno podes deixar o que pus aqui
	newCall := downstreamArgs.(Call)
	newNodeIdx := len(crdt.NodeArr)
	isNoOp := false
	// for v ∈ V
	for i := range crdt.NodeArr {
		compVal := crdt.NodeArr[i].Value.Time.Compare(newCall.Time)
		// if v ≺ c then
		//fmt.Println("Checking:", crdt.NodeArr[i].Value.Op, newCall.Op, compVal)			//DEBUG PRINT
		if compVal < 0 {
			// E <- E ∪ {<v,c>}
			(&crdt.NodeArr[i]).addEdge(newNodeIdx)
			// else if v ∥ c ∧ opsConflict(v, c) then (v and c are concurrent and callTypeoperation is among blocked operations)
		} else if compVal == clocksi.ConcurrentTs {
			//fmt.Println("CONFLICT DETECTED:", crdt.NodeArr[i].Value.Op, ",", newCall.Op)			//DEBUG PRINT
			if crdt.NodeArr[i].Value.Blocks(&newCall) { //new Call is a No Op
				//mark call as no-opl
				isNoOp = true
			}
			if newCall.Blocks(&crdt.NodeArr[i].Value) { //Existing Call is a No Op
				crdt.NodeArr[i].IsNoOp = true
			}
		}
	}
	//V ← V ∪ {c}
	crdt.NodeArr = append(crdt.NodeArr, Node{Value: newCall, Edges: nil, IsNoOp: isNoOp})

	var effectV Effect = NoEffect{}
	return &effectV
}

func (crdt *NoOpCrdt) IsOperationWellTyped(args UpdateArguments) (ok bool, err error) {
	return true, nil
}

func (crdt *NoOpCrdt) Copy() (copyCRDT InversibleCRDT) {
	newCRDT := NoOpCrdt{
		CRDTVM:       crdt.CRDTVM.copy(),
		DataContent: crdt.DataContent.Copy(),
		NodeArr:      make([]Node, len(crdt.NodeArr)),
		//Adicionar outros campos que pertençam ao NoOpCrdt
	}
	for i, node := range crdt.NodeArr {
		newCRDT.NodeArr[i] = node.Copy()
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
