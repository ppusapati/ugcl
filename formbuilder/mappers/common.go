package mappers

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// StringToAny converts a string value to protobuf Any
func StringToAny(value string) (*anypb.Any, error) {
	strValue := &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: value}}
	return anypb.New(strValue)
}

// InterfaceToAny converts an interface{} value to protobuf Any
func InterfaceToAny(value interface{}) (*anypb.Any, error) {
	if value == nil {
		return nil, nil
	}
	
	switch v := value.(type) {
	case string:
		strValue := &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: v}}
		return anypb.New(strValue)
	case int:
		numValue := &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}
		return anypb.New(numValue)
	case int32:
		numValue := &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}
		return anypb.New(numValue)
	case int64:
		numValue := &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}
		return anypb.New(numValue)
	case float32:
		numValue := &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}
		return anypb.New(numValue)
	case float64:
		numValue := &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: v}}
		return anypb.New(numValue)
	case bool:
		boolValue := &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: v}}
		return anypb.New(boolValue)
	default:
		// Try to serialize as JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value to JSON: %w", err)
		}
		strValue := &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: string(jsonBytes)}}
		return anypb.New(strValue)
	}
}

// AnyToInterface converts protobuf Any to interface{}
func AnyToInterface(anyValue *anypb.Any) (interface{}, error) {
	if anyValue == nil {
		return nil, nil
	}

	var value structpb.Value
	if err := anyValue.UnmarshalTo(&value); err != nil {
		return anyValue.String(), nil // Fallback to string representation
	}

	switch v := value.Kind.(type) {
	case *structpb.Value_StringValue:
		return v.StringValue, nil
	case *structpb.Value_NumberValue:
		return v.NumberValue, nil
	case *structpb.Value_BoolValue:
		return v.BoolValue, nil
	case *structpb.Value_NullValue:
		return nil, nil
	default:
		return anyValue.String(), nil
	}
}

// MapStringInterfaceToProtoAny converts map[string]interface{} to map[string]*anypb.Any
func MapStringInterfaceToProtoAny(m map[string]interface{}) (map[string]*anypb.Any, error) {
	result := make(map[string]*anypb.Any)
	for key, value := range m {
		anyValue, err := InterfaceToAny(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for key %s: %w", key, err)
		}
		result[key] = anyValue
	}
	return result, nil
}

// ProtoAnyToMapStringInterface converts map[string]*anypb.Any to map[string]interface{}
func ProtoAnyToMapStringInterface(m map[string]*anypb.Any) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for key, anyValue := range m {
		value, err := AnyToInterface(anyValue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for key %s: %w", key, err)
		}
		result[key] = value
	}
	return result, nil
}
