package crdt

import (
	"potionDB/crdt/clocksi"
	"potionDB/crdt/proto"
)

// Operations
type Operation interface {
	OpEqual(Operation)bool
	Copy()Operation
	Precondition(state State) bool
	BlockGenerator() []Operation
	Process(state State)
	GetCRDTType() proto.CRDTType
	String() string

	//for protobuf
	FromUpdateObject(*proto.ApbUpdateOperation) UpdateArguments
	ToUpdateObject() *proto.ApbUpdateOperation
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

func (call *Call) Copy()Call {
	newCall := Call{
		Op: call.Op.Copy(),
		BlockedOps: make([]Operation, len(call.BlockedOps)),
		Time : call.Time.Copy(),
	}
	for i, blockOp := range call.BlockedOps {
		newCall.BlockedOps[i] = blockOp.Copy()
	}
	return newCall
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
func (call Call) MustReplicate() bool {return false}