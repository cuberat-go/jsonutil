package jsonutil_test

import (
	// Core/builtin modules.

	"bytes"
	"encoding/json"
	"testing"

	// Third-party modules.
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	// Generated protobuf code.
	"github.com/cuberat-go/jsonutil/test/proto_stuff/my_proto_stuff"

	// First-party modules.
	"github.com/cuberat-go/jsonutil"
)

func TestEncoder(t *testing.T) {
	expectedOutStr := `{"field1":"value1","field2":42}`
	var expected any
	err := json.Unmarshal([]byte(expectedOutStr), &expected)
	if err != nil {
		t.Fatalf("Failed to unmarshal expected output: %v", err)
	}

	protoEncode := func(m proto.Message) ([]byte, error) {
		return protojson.MarshalOptions{UseProtoNames: true}.Marshal(m)
	}

	b := &bytes.Buffer{}

	myProto := &my_proto_stuff.MyProtoStuff{Field1: "value1", Field2: 42}
	enc := jsonutil.NewEncoder(b).WithMarshalers(
		jsonutil.JoinMarshalers(
			jsonutil.MarshalFunc(protoEncode),
		))
	err = enc.Encode(myProto)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	got := b.Bytes()
	t.Logf("Encoded output: %s", string(got))

	var gotAny any
	err = json.Unmarshal(got, &gotAny)
	if err != nil {
		t.Fatalf("Failed to unmarshal encoded output: %v", err)
	}
	assert.Equal(t, expected, gotAny)
}

func TestMarshaler(t *testing.T) {
	expectedOutStr := `{"field1":"value1","field2":42}`
	var expected any
	err := json.Unmarshal([]byte(expectedOutStr), &expected)
	if err != nil {
		t.Fatalf("Failed to unmarshal expected output: %v", err)
	}

	protoEncode := func(m proto.Message) ([]byte, error) {
		return protojson.MarshalOptions{UseProtoNames: true}.Marshal(m)
	}

	myProto := &my_proto_stuff.MyProtoStuff{Field1: "value1", Field2: 42}
	enc := jsonutil.WithMarshalers(
		jsonutil.JoinMarshalers(
			jsonutil.MarshalFunc(protoEncode),
		))
	got, err := enc.Marshal(myProto)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	t.Logf("Encoded output: %s", string(got))

	var gotAny any
	err = json.Unmarshal(got, &gotAny)
	if err != nil {
		t.Fatalf("Failed to unmarshal encoded output: %v", err)
	}
	assert.Equal(t, expected, gotAny)
}

func TestProtoWithOneof(t *testing.T) {
	expectedOutStr := `{"top_field":"value1","field2":42}`
	var expected any
	err := json.Unmarshal([]byte(expectedOutStr), &expected)
	if err != nil {
		t.Fatalf("Failed to unmarshal expected output: %v", err)
	}

	protoEncode := func(m proto.Message) ([]byte, error) {
		return protojson.MarshalOptions{UseProtoNames: true}.Marshal(m)
	}

	myProto := &my_proto_stuff.ProtoWithOneof{
		TopField:  "value1",
		TestOneof: &my_proto_stuff.ProtoWithOneof_Field2{Field2: 42},
	}
	enc := jsonutil.WithMarshalers(
		jsonutil.MarshalFunc(protoEncode),
	)
	got, err := enc.Marshal(myProto)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	t.Logf("Encoded output: %s", string(got))

	builtin_got, _ := json.Marshal(myProto)
	t.Logf("Builtin encoded output: %s", string(builtin_got))

	var gotAny any
	err = json.Unmarshal(got, &gotAny)
	if err != nil {
		t.Fatalf("Failed to unmarshal encoded output: %v", err)
	}
	assert.Equal(t, expected, gotAny)
}

func TestStruct(t *testing.T) {
	expectedOutStr := `{"field1":"value1","field2":42,"field3":{"top_field":"value3","field2":42}}`
	var expected any
	err := json.Unmarshal([]byte(expectedOutStr), &expected)
	if err != nil {
		t.Fatalf("Failed to unmarshal expected output: %v", err)
	}

	protoEncode := func(m proto.Message) ([]byte, error) {
		return protojson.MarshalOptions{UseProtoNames: true}.Marshal(m)
	}
	type MyStruct struct {
		Field1 string                         `json:"field1"`
		Field2 int                            `json:"field2"`
		Field3 *my_proto_stuff.ProtoWithOneof `json:"field3"`
	}

	myStruct := &MyStruct{
		Field1: "value1",
		Field2: 42,
		Field3: &my_proto_stuff.ProtoWithOneof{
			TopField:  "value3",
			TestOneof: &my_proto_stuff.ProtoWithOneof_Field2{Field2: 42},
		},
	}
	enc := jsonutil.WithMarshalers(
		jsonutil.JoinMarshalers(
			jsonutil.MarshalFunc(protoEncode),
		))
	got, err := enc.Marshal(myStruct)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	t.Logf("Encoded output: %s", string(got))

	var gotAny any
	err = json.Unmarshal(got, &gotAny)
	if err != nil {
		t.Fatalf("Failed to unmarshal encoded output: %v", err)
	}
	assert.Equal(t, expected, gotAny)
}

func TestTags(t *testing.T) {
	expectedOutStr := `{"field1":"value1","field2":42,"Field4FieldName":"should be included"}`
	var expected any
	err := json.Unmarshal([]byte(expectedOutStr), &expected)
	if err != nil {
		t.Fatalf("Failed to unmarshal expected output: %v", err)
	}

	type MyStruct struct {
		Field1Lower     string `json:"field1"`
		Field2Lower     int    `json:"field2"`
		Field3Ignore    string `json:"-"`
		Field4FieldName string `json:","`
		Field5NoOutput  string `json:"field5,omitempty"`
	}

	myStruct := &MyStruct{
		Field1Lower:     "value1",
		Field2Lower:     42,
		Field3Ignore:    "should be ignored",
		Field4FieldName: "should be included",
		Field5NoOutput:  "",
	}

	got, err := jsonutil.Marshal(myStruct)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	t.Logf("Encoded output: %s", string(got))

	var gotAny any
	err = json.Unmarshal(got, &gotAny)
	if err != nil {
		t.Fatalf("Failed to unmarshal encoded output: %v", err)
	}
	assert.Equal(t, expected, gotAny)

}
