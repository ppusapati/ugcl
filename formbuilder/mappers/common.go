package mappers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
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

	fmt.Printf("DEBUG: Processing Any value - TypeUrl: %s, Value length: %d\n", anyValue.TypeUrl, len(anyValue.Value))

	var value structpb.Value
	if err := proto.Unmarshal(anyValue.Value, &value); err != nil {
		fmt.Printf("DEBUG: Unmarshal failed: %v\n", err)

		// If TypeUrl is empty and we have raw bytes, try different approaches
		if anyValue.TypeUrl == "" && len(anyValue.Value) > 0 {
			rawBytes := anyValue.Value
			rawString := string(rawBytes)
			fmt.Printf("DEBUG: Raw bytes as string: '%s'\n", rawString)
			fmt.Printf("DEBUG: Raw bytes hex: %x\n", rawBytes)

			// Try to parse as protobuf Value directly from bytes
			var directValue structpb.Value
			if err := proto.Unmarshal(rawBytes, &directValue); err == nil {
				fmt.Printf("DEBUG: Direct protobuf unmarshal successful\n")
				switch v := directValue.Kind.(type) {
				case *structpb.Value_StringValue:
					fmt.Printf("DEBUG: Direct string value: %s\n", v.StringValue)
					return v.StringValue, nil
				case *structpb.Value_NumberValue:
					fmt.Printf("DEBUG: Direct number value: %f\n", v.NumberValue)
					return v.NumberValue, nil
				case *structpb.Value_BoolValue:
					fmt.Printf("DEBUG: Direct bool value: %t\n", v.BoolValue)
					return v.BoolValue, nil
				}
			} else {
				fmt.Printf("DEBUG: Direct protobuf unmarshal failed: %v\n", err)
			}

			// Check if it looks like a valid UTF-8 string (printable characters)
			isPrintable := true
			for _, r := range rawString {
				if r < 32 || r > 126 { // Not printable ASCII
					isPrintable = false
					break
				}
			}

			if isPrintable && len(rawString) > 0 {
				fmt.Printf("DEBUG: Using raw string: '%s'\n", rawString)
				return rawString, nil
			}

			// Fallback: try base64 decoding
			fmt.Printf("DEBUG: Attempting base64 decode of: %s\n", rawString)
			if decoded, err := base64.StdEncoding.DecodeString(rawString); err == nil {
				fmt.Printf("DEBUG: Base64 decode successful: %s\n", string(decoded))
				return string(decoded), nil
			} else {
				fmt.Printf("DEBUG: Base64 decode failed: %v\n", err)
			}
		}

		// Extract clean string from protobuf representation
		str := anyValue.String()
		if strings.HasPrefix(str, `value:"`) && strings.HasSuffix(str, `"`) {
			// Extract the actual value between value:" and "
			cleanValue := str[7 : len(str)-1] // Remove 'value:"' and '"'
			fmt.Printf("DEBUG: Extracted clean value: %s\n", cleanValue)
			return cleanValue, nil
		}
		fmt.Printf("DEBUG: Falling back to anyValue.String(): %s\n", str)
		return str, nil // Final fallback to string representation
	}

	fmt.Printf("DEBUG: Unmarshal successful, processing value type\n")
	switch v := value.Kind.(type) {
	case *structpb.Value_StringValue:
		fmt.Printf("DEBUG: String value: %s\n", v.StringValue)
		return v.StringValue, nil
	case *structpb.Value_NumberValue:
		fmt.Printf("DEBUG: Number value: %f\n", v.NumberValue)
		return v.NumberValue, nil
	case *structpb.Value_BoolValue:
		fmt.Printf("DEBUG: Bool value: %t\n", v.BoolValue)
		return v.BoolValue, nil
	case *structpb.Value_NullValue:
		fmt.Printf("DEBUG: Null value\n")
		return nil, nil
	default:
		fmt.Printf("DEBUG: Unknown value type, falling back to string: %s\n", anyValue.String())
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
