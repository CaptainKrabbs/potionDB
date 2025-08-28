package crdt

import (
	"encoding/binary"
	"fmt"
	"potionDB/crdt/clocksi"
	"potionDB/crdt/proto"

	pb "google.golang.org/protobuf/proto"
)

// Operations
type Operation interface {
	UpdateArguments
	ProtoUpd
	OpEqual(Operation) bool
	Copy() Operation
	Precondition(state State) bool
	BlockGenerator() []Operation
	Process(state NoOpState) NoOpState
	String() string    //Returns formatted string of name and parameters
	GetOpName() string //Returns formatted string of name

	//for protobuf
	GetStateCode() int32
	GetOpCode() int32
	GetNumParams() int
	GetSerializedParams() [][]byte
}

type OperationAbstract struct {
}

// Auxiliary method for ToUpdateObject that does the entire logic since method logic is always the same
func ToUpdateObjectFrame(a Operation) (protobuf *proto.ApbUpdateOperation) {
	stateCode := a.GetStateCode()
	opCode := a.GetOpCode()
	params := a.GetSerializedParams()
	return &proto.ApbUpdateOperation{Noop: &proto.ApbNoOpUpdate{StateCode: &stateCode, OpCode: &opCode, Params: params}}
}

/* Auxiliary method for FromUpdateObject
 * Verifies correctness of codes for a protobuf attempting to be converted to an Operation of the given type.
*/
func VerifyOpProtobuf(a Operation, protobuf *proto.ApbUpdateOperation) (ok bool) {
	//Fail conditions: wrong
	if a.GetStateCode() != *protobuf.GetNoop().StateCode {
		fmt.Printf("Error occurred: Invalid State type for Operation %v. Expected: %v Given: %v", a.GetOpName(), a.GetStateCode(), protobuf.GetNoop().StateCode)
	} else if a.GetOpCode() != *protobuf.GetNoop().OpCode {
		fmt.Printf("Error occurred: Invalid Operation code for Operation %v. Expected: %v Given: %v", a.GetOpName(), a.GetOpCode(), protobuf.GetNoop().OpCode)
	} else if a.GetNumParams() != len(protobuf.GetNoop().GetParams()) {
		fmt.Printf("Error occurred: Invalid number of parameters for Operation %v. Expected: %v Given: %v", a.GetOpName(), a.GetNumParams(), len(protobuf.GetNoop().Params))
	} else {
		return true
	}
	return false
}

// Message struct that stores an operation and those it blocks
type Message struct {
	Op         Operation
	BlockedOps []Operation
}

func (msg Message) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (msg Message) MustReplicate() bool         { return true }

// Converts a message to a call struct given a vector Timestamp
func (msg Message) ToCall(clock clocksi.Timestamp) Call {
	return Call{Op: msg.Op, BlockedOps: msg.BlockedOps, Clock: clock}
}

//end of aux functions to convert clocks


func (msg Message) FromReplicatorObj(protobuf *proto.ProtoOpDownstream) DownstreamArguments {
	return msg.protoToMessage(protobuf)
}

func (msg Message) ToReplicatorObj() (protobuf *proto.ProtoOpDownstream) {
	blocks := make([]*proto.ApbNoOpUpdate, len(msg.BlockedOps))
	for i, op := range msg.BlockedOps {
		blocks[i] = op.ToUpdateObject().Noop
	}
	return &proto.ProtoOpDownstream{NoOpOp: &proto.ProtoNoOpDownstream{Op: msg.Op.ToUpdateObject().Noop, Blocks: blocks}}
}

// Aux function to get individual op
func opFromProtoOpDownstream(protobuf *proto.ApbNoOpUpdate) (op Operation) {
	op, ok := updateNoOpProtoToAntidoteUpdate(&proto.ApbUpdateOperation{Noop: protobuf}).(Operation)
	if !ok {
		fmt.Println("[noOpOperation][ERROR] Object returned by proto to Operation conversion doesn't have Operation type")
		//Should never happen
		return &NoOp{}
	}
	return op
}

func (msg *Message) protoToMessage(protobuf *proto.ProtoOpDownstream) Message {
	op := opFromProtoOpDownstream(protobuf.GetNoOpOp().GetOp())
	var blocks []Operation

	switch op.(type) {
	case *NoOp:
		blocks = nil
	default:
		blocks = make([]Operation, len(protobuf.GetNoOpOp().GetBlocks()))
		for i, block := range protobuf.GetNoOpOp().GetBlocks() {
			blocks[i] = opFromProtoOpDownstream(block)
		}
	}
	return Message{Op: op, BlockedOps: blocks}
}

// struct that stores an operation, those it blocks and a vector clock Timestamp
type Call struct {
	Op         Operation
	BlockedOps []Operation
	Clock      clocksi.Timestamp
}

func (call *Call) Copy() Call {
	newCall := Call{
		Op:         call.Op.Copy(),
		BlockedOps: make([]Operation, len(call.BlockedOps)),
		Clock:      call.Clock.Copy(),
	}
	for i, blockOp := range call.BlockedOps {
		newCall.BlockedOps[i] = blockOp.Copy()
	}
	return newCall
}

// check if any blocking operation matches the one in the call.
//
// if so early return true.
// otherwise false.
func (actualCall *Call) Blocks(otherCall *Call) bool {
	for i := range actualCall.BlockedOps {
		if (actualCall.BlockedOps[i]).OpEqual(otherCall.Op) {
			return true
		}
	}
	return false
}

// Aux functions to convert clocks
func clocksiToProtoClock(ts *clocksi.ClockSiTimestamp) (protoClock *proto.ProtoClock) {
	entries := make([]*proto.ProtoStableClock, len(ts.VectorClock))
	i := 0
	for senderID, replicaTs := range ts.VectorClock {
		entries[i] = &proto.ProtoStableClock{SenderID: pb.Int32(int32(senderID)), ReplicaTs: &replicaTs}
		i++
	}
	return &proto.ProtoClock{Entries: entries}
}

func clocksiFromProtoClock(protoClock *proto.ProtoClock) (ts *clocksi.ClockSiTimestamp) {
	vectorClock := map[int16]int64{}
	for _, clock := range protoClock.Entries {
		vectorClock[int16(*clock.SenderID)] = *clock.ReplicaTs
	}
	return &clocksi.ClockSiTimestamp{VectorClock: vectorClock}
}

//end of aux functions to convert clocks

func (call *Call) CallToProto() (protobuf *proto.ProtoNoOpCall) {
	convClock, ok := call.Clock.(clocksi.ClockSiTimestamp)
	if !ok {
		return nil
	}

	clock := clocksiToProtoClock(&convClock)
	blocks := make([]*proto.ApbNoOpUpdate, len(call.BlockedOps))
	for i, op := range call.BlockedOps {
		blocks[i] = op.ToUpdateObject().Noop
	}
	return &proto.ProtoNoOpCall{Op: call.Op.ToUpdateObject().Noop, Blocks: blocks, Clock: clock}
}

func (call *Call) ProtoToCall(protobuf *proto.ProtoNoOpCall) Call {
	op := opFromProtoOpDownstream(protobuf.GetOp())
	clock := clocksiFromProtoClock(protobuf.GetClock())
	var blocks []Operation

	switch op.(type) {
	case *NoOp:
		blocks = nil
	default:
		blocks = make([]Operation, len(protobuf.GetBlocks()))
		for i, block := range protobuf.GetBlocks() {
			blocks[i] = opFromProtoOpDownstream(block)
		}
	}
	return Call{Op: op, BlockedOps: blocks, Clock: clock}
}

func (call Call) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (call Call) MustReplicate() bool         { return false }

//----------NoOp

var (
	NoOpCode      int32 = 0
	NoOpNumParams       = 0
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

func (a *NoOp) Process(state NoOpState) NoOpState {
	return state
}

func (a *NoOp) String() string {
	return "Operation: NoOp"
}

func (a *NoOp) GetOpName() string {
	return "NoOp"
}

func (a *NoOp) GetStateCode() (num int32) {
	return DecisionStateCode
}

func (a *NoOp) GetOpCode() (num int32) {
	return NoOpCode
}

func (a *NoOp) GetNumParams() (num int) {
	return NoOpNumParams
}

func (a *NoOp) GetSerializedParams() [][]byte {
	return [][]byte{}
}

func (a *NoOp) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return ToUpdateObjectFrame(a)
}

func (a *NoOp) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if !VerifyOpProtobuf(a, protobuf) {
		return NoOp{}
	}
	return NoOp{}
}

//----------DetermineState

var (
	DetermineStateOpCode    int32 = 1
	DetermineStateNumParams       = 1
	NewStateCodeParamIdx          = 0
)

type DetermineStateOp struct {
	NewStateCode int32
}

func (args *DetermineStateOp) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }

// Methods for NoOp, the rest of the operations are in noOpMusicApp.go
func (a *DetermineStateOp) OpEqual(o Operation) bool {
	oConv, ok := o.(*DetermineStateOp)
	return ok && a.NewStateCode == oConv.NewStateCode
}

func (a *DetermineStateOp) Copy() Operation {
	return &DetermineStateOp{NewStateCode: a.NewStateCode}
}

func (a *DetermineStateOp) Precondition(state State) bool {
	return true
}

func (a *DetermineStateOp) BlockGenerator() []Operation {
	return nil
}

func (a *DetermineStateOp) Process(state NoOpState) NoOpState {
	switch a.NewStateCode {
	case (&MusicState{}).GetStateCode():
		newState := &MusicState{}
		newState.Initialize()
		return newState
	}
	return state
}

func (a *DetermineStateOp) String() string {
	return "Operation: DetermineState"
}

func (a *DetermineStateOp) GetOpName() string {
	return "DetermineState"
}

func (a *DetermineStateOp) GetStateCode() (num int32) {
	return DecisionStateCode
}

func (a *DetermineStateOp) GetOpCode() (num int32) {
	return DetermineStateOpCode
}

func (a *DetermineStateOp) GetNumParams() (num int) {
	return DetermineStateNumParams
}

func (a *DetermineStateOp) GetSerializedParams() [][]byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:4], uint32(a.NewStateCode))

	var bytes = make([][]byte, DetermineStateNumParams)
	bytes[NewStateCodeParamIdx] = buf
	return bytes
}

func (a *DetermineStateOp) ToUpdateObject() (protobuf *proto.ApbUpdateOperation) {
	return ToUpdateObjectFrame(a)
}

func (a *DetermineStateOp) FromUpdateObject(protobuf *proto.ApbUpdateOperation) (op UpdateArguments) {
	if !VerifyOpProtobuf(a, protobuf) {
		return NoOp{}
	}
	return &DetermineStateOp{NewStateCode: int32(binary.BigEndian.Uint32(protobuf.GetNoop().GetParams()[NewStateCodeParamIdx]))}
}
