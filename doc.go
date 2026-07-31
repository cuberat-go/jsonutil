// Package jsonutil provides utilities for JSON encoding with support for custom
// marshalers. This package duplicates functionality in the experimental
// encoding/json/v2 package and will be deprecated once the built-in package
// provides equivalent functionality.
//
// Custom marshalers are handy for encoding types that do not have built-in JSON
// support or require special handling, e.g., protobufs with oneofs or enums.
// E.g., for the protobuf message
//
//	message ProtoWithOneof {
//	    string top_field = 1;
//	    oneof test_oneof {
//	        string field1 = 2;
//	        int32 field2 = 3;
//	    }
//	}
//
// The Marshal call to the core encoding/json module outputs something like the
// following:
//
//	{"top_field":"value1","TestOneof":{"Field2":42}}
//
// As for the jsonutil package with custom marshalers:
//
//	myProto := &my_proto_stuff.ProtoWithOneof{
//	    TopField:  "value1",
//	    TestOneof: &my_proto_stuff.ProtoWithOneof_Field2{Field2: 42},
//	}
//
//	protoEncode := func(m proto.Message) ([]byte, error) {
//	    return protojson.MarshalOptions{UseProtoNames: true}.Marshal(m)
//	}
//	got, err := jsonutil.WithMarshalers(
//	    jsonutil.MarshalFunc(protoEncode),
//	).Marshal(myProto)
//	fmt.Printf("Encoded output: %s\n", string(got))
//
//	{"top_field":"value1","field2":42}
//
// Use the JoinMarshalers function to combine multiple marshalers into a single
// one to pass to the WithMarshalers function. The first marshalers in the
// list are tried first, and if they return the Skip error, the next
// marshalers are tried until one succeeds or until a marshaler returns a
// non-Skip error. If all marshalers return Skip, the default encoding/json
// module is used to encode the value.
package jsonutil
