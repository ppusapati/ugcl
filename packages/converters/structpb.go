package converters

import (
	"encoding/json"

	"google.golang.org/protobuf/types/known/structpb"
)

func MapToStructPB(m map[string]any) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		panic("failed to convert map to structpb.Struct: " + err.Error())
	}
	return s
}

func StructPBToMap(s *structpb.Struct) map[string]any {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

func StructPBToJSON(s *structpb.Struct) []byte {
	if s == nil {
		return nil
	}
	b, _ := json.Marshal(s.AsMap())
	return b
}

func JSONToStructPB(b []byte) *structpb.Struct {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	s, _ := structpb.NewStruct(m)
	return s
}

func IsEmptyStructPB(s *structpb.Struct) bool {
	return s == nil || len(s.Fields) == 0
}
