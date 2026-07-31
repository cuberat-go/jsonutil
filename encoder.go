package jsonutil

import (
	// Core/builtin modules.
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var SkipFunc = errors.New("json: skip function")

// A custom marshaler or unmarshaller for a specific type.
type arshaler struct {
	Type reflect.Type
	Func any
}

// Structure used for managing custom marshalers.
type Marshalers struct {
	arshalers []*arshaler
}

// Return any custom marshalers applicable to the given type.
func (m *Marshalers) lookup(t reflect.Type) []*arshaler {
	if m == nil {
		return nil
	}

	var result []*arshaler
	for _, ar := range m.arshalers {
		if castableTo(t, ar.Type) {
			result = append(result, ar)
		}
	}
	return result
}

// Determines if a value of one type can be cast to another type.
// This is used to check if a custom marshaler is applicable to a given type.
func castableTo(from, to reflect.Type) bool {
	switch to.Kind() {
	case reflect.Interface:
		if from.Kind() == reflect.Pointer {
			return from.Implements(to)
		}
		return reflect.PointerTo(from).Implements(to)
	case reflect.Pointer:
		if from.Kind() == reflect.Pointer {
			return from == to
		}
		return reflect.PointerTo(from) == to
	default:
		return from == to
	}
}

// Joins multiple Marshalers into a single Marshalers instance.
func JoinMarshalers(ms ...*Marshalers) *Marshalers {
	joined := &Marshalers{}
	for _, m := range ms {
		joined.arshalers = append(joined.arshalers, m.arshalers...)
	}
	return joined
}

// Creates a Marshalers instance from a specific marshaling function for a given
// type.
func MarshalFunc[T any](fn func(T) ([]byte, error)) *Marshalers {
	return &Marshalers{
		arshalers: []*arshaler{
			{
				Type: reflect.TypeFor[T](),
				Func: func(vAny any) ([]byte, error) {
					if v, ok := vAny.(T); ok {
						return fn(v)
					}
					return nil, SkipFunc
				},
			},
		},
	}
}

// Structure representing a JSON encoder with support for custom marshalers.
type Encoder struct {
	marshalers *Marshalers
}

func newEncoder() *Encoder {
	return &Encoder{}
}

// Marshals a value using the default encoder with no custom marshalers.
func Marshal(v any) ([]byte, error) {
	e := newEncoder()
	return e.Marshal(v)
}

// Marshals a value using the encoder, considering any custom marshalers that
// have been set with the WithMarshalers method.
func (e *Encoder) Marshal(v any) ([]byte, error) {
	d, err := e.marshal(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(d)
}

func (e *Encoder) marshal(v any) (json.RawMessage, error) {
	val := reflect.ValueOf(v)

	if e.marshalers == nil {
		d, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(d), nil
	}

	for _, ar := range e.marshalers.lookup(val.Type()) {
		if fn, ok := ar.Func.(func(any) ([]byte, error)); ok {
			if data, err := fn(v); err == nil {
				return data, nil
			} else if err != SkipFunc {
				return nil, err
			}
		}
	}

	for val.Kind() == reflect.Interface || val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		return e.marshalStruct(v)
	case reflect.Array, reflect.Slice:
		return e.marshalSlice(v)
	case reflect.Map:
		return e.marshalMap(v)
	default:
		d, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(d), nil
	}
}

func (e *Encoder) marshalStruct(v any) (json.RawMessage, error) {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Interface || val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("marshalStruct expects a struct, got %v",
			val.Kind())
	}

	newMap := make(map[string]json.RawMessage, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if !field.IsExported() {
			continue
		}
		fieldName := field.Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			fieldName = strings.Split(jsonTag, ",")[0]
		}
		if fieldName == "-" || fieldName == "" {
			continue
		}
		fieldValue := val.Field(i).Interface()
		d, err := e.marshal(fieldValue)
		if err != nil {
			return nil, err
		}
		newMap[fieldName] = d
	}

	newMapBytes, err := json.Marshal(newMap)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(newMapBytes), nil
}

func (e *Encoder) marshalSlice(v any) (json.RawMessage, error) {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Interface || val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Array && val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("marshalSlice expects an array or slice, "+
			"got %v", val.Kind())
	}

	newSlice := make([]json.RawMessage, val.Len())

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i).Interface()
		d, err := e.marshal(elem)
		if err != nil {
			return nil, err
		}
		newSlice[i] = d
	}

	newSliceBytes, err := json.Marshal(newSlice)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(newSliceBytes), nil
}

func (e *Encoder) marshalMap(v any) (json.RawMessage, error) {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Interface || val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("marshalMap expects a map, got %v", val.Kind())
	}

	newMap := make(map[string]json.RawMessage, val.Len())
	for _, key := range val.MapKeys() {
		mapValue := val.MapIndex(key).Interface()
		d, err := e.marshal(mapValue)
		if err != nil {
			return nil, err
		}
		newMap[key.String()] = d
	}

	newMapBytes, err := json.Marshal(newMap)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(newMapBytes), nil
}

// Sets the custom marshalers for the encoder and returns the encoder itself.
func (e *Encoder) WithMarshalers(marshalers *Marshalers) *Encoder {
	e.marshalers = marshalers
	return e
}

// Creates a new encoder with the specified custom marshalers.
func WithMarshalers(marshalers *Marshalers) *Encoder {
	return newEncoder().WithMarshalers(marshalers)
}
