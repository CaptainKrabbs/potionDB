/*
Contains the definition of the NoOpState interface
Contains the implementation of the placeholder state
used on creation before formal NoOpState attribution via Operation
(only Operation available in DecisionState)
*/
package crdt

import (
	"fmt"
	"potionDB/crdt/proto"
)

var DecisionStateCode int32 = 0

type NoOpState interface {
	State
	ProtoState
	Initialize()
	Copy() NoOpState
	GetStateCode() int32

	Serialize() [][]byte
	Deserialize([][]byte) NoOpState
}

/*
A placeholder initial state for NoOpCrdt.
Used prior to assigning proper state, allows for generic NoOpState implementation without
nil pointer dereferencing errors
*/
type DecisionState struct {
}

func (s *DecisionState) GetStateCode() int32 { return DecisionStateCode }

func (s *DecisionState) GetCRDTType() proto.CRDTType { return proto.CRDTType_NOOP }
func (s *DecisionState) GetREADType() proto.READType { return proto.READType_FULL }

func (s *DecisionState) Initialize() {
}

func (s *DecisionState) Copy() NoOpState {
	return &DecisionState{}
}

func (s *DecisionState) FromUpdateObject(protobuf *proto.ApbUpdateOperation) UpdateArguments {
	if protobuf.GetNoop().GetStateCode() == s.GetStateCode() {
		switch protobuf.GetNoop().GetOpCode() {
		case (&DetermineStateOp{}).GetOpCode():
			return (&DetermineStateOp{}).FromUpdateObject(protobuf)
		}
	}
	return nil
}

// Aux function for ToReadResp to avoid repeating code
func ToReadRespFrame(s NoOpState) (protobuf *proto.ApbReadObjectResp) {
	stateCode := s.GetStateCode()
	return &proto.ApbReadObjectResp{Noop: &proto.ApbGetNoOpResp{StateCode: &stateCode, StateData: s.Serialize()}}
}

func (s *DecisionState) ToReadResp() (protobuf *proto.ApbReadObjectResp) {
	return ToReadRespFrame(s)
}

func (s *DecisionState) FromReadResp(protobuf *proto.ApbReadObjectResp) State {
	return s.Deserialize(protobuf.GetNoop().GetStateData())
}

func (s *DecisionState) Serialize() [][]byte {
	return [][]byte{}
}

func (s *DecisionState) Deserialize(bytes [][]byte) NoOpState {
	return &DecisionState{}
}

// General functions

func GetNoOpStateFromProto(protobuf *proto.ProtoNoOpState) NoOpState {
	switch protobuf.GetStateCode() {
	case DecisionStateCode:
		return (&DecisionState{}).Deserialize(protobuf.GetStateData())
	case MusicStateCode:
		return (&MusicState{}).Deserialize(protobuf.GetStateData())
	default:
		fmt.Printf("[Error][noOpMusicAppState] StateCode from ProtoNoOpState not valid for a NoOpCrdt: %v\n", protobuf.GetStateCode())
		return nil
	}
}
