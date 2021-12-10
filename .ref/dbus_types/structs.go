package dbus_types

// https://dbus.freedesktop.org/doc/dbus-specification.html#type-system

type DbusType struct {
	TypeName  string
	Symbol    rune
	Desc      string
	ZeroValue interface{}
}

/*
	BASIC TYPES
*/

var DbusByte DbusType = DbusType{
	TypeName:  "BYTE",
	Symbol:    'y',
	Desc:      "Unsigned 8-bit integer",
	ZeroValue: byte(0x0),
}

var DbusBoolean DbusType = DbusType{
	TypeName:  "BOOLEAN",
	Symbol:    'b',
	Desc:      "Boolean value: 0 is false, 1 is true, any other value allowed by the marshalling format is invalid",
	ZeroValue: false,
}

var DbusInt16 DbusType = DbusType{
	TypeName:  "INT16",
	Symbol:    'n',
	Desc:      "Signed (two's complement) 16-bit integer",
	ZeroValue: int16(0),
}

var DbusUint16 DbusType = DbusType{
	TypeName:  "UINT16",
	Symbol:    'q',
	Desc:      "Unsigned 16-bit integer",
	ZeroValue: uint16(0),
}

var DbusInt32 DbusType = DbusType{
	TypeName:  "INT32",
	Symbol:    'i',
	Desc:      "Signed (two's complement) 32-bit integer",
	ZeroValue: int32(0),
}

var DbusUint32 DbusType = DbusType{
	TypeName:  "UINT32",
	Symbol:    'u',
	Desc:      "Unsigned 32-bit integer",
	ZeroValue: uint32(0),
}

var DbusInt64 DbusType = DbusType{
	TypeName:  "INT64",
	Symbol:    'x',
	Desc:      "Signed (two's complement) 64-bit integer (mnemonic: x and t are the first characters in \"sixty\" not already used for something more common)",
	ZeroValue: int64(0),
}

var DbusUint64 DbusType = DbusType{
	TypeName:  "UINT64",
	Symbol:    't',
	Desc:      "Unsigned 64-bit integer",
	ZeroValue: uint64(0),
}

var DbusDoubleFloat DbusType = DbusType{
	TypeName:  "DOUBLE",
	Symbol:    'd',
	Desc:      "IEEE 754 double-precision floating point",
	ZeroValue: float64(0),
}

var DbusUnixFD DbusType = DbusType{
	TypeName:  "UNIX_FD",
	Symbol:    'h',
	Desc:      "Unsigned 32-bit integer representing an index into an out-of-band array of file descriptors, transferred via some platform-specific mechanism (mnemonic: h for handle)",
	ZeroValue: uint32(0), // See https://pkg.go.dev/github.com/godbus/dbus#UnixFDIndex
}

var DbusString DbusType = DbusType{
	TypeName:  "STRING",
	Symbol:    'o',
	Desc:      "No extra constraints",
	ZeroValue: "",
}

var DbusObjectPath DbusType = DbusType{
	TypeName:  "OBJECT_PATH",
	Symbol:    'o',
	Desc:      "A syntactically valid Path for Dbus",
	ZeroValue: nil, // ???
}

var DbusSignature DbusType = DbusType{
	TypeName:  "SIGNATURE",
	Symbol:    'g',
	Desc:      "0 or more single complete types", // ???
	ZeroValue: nil,                               // ???
}

/*
	CONTAINER TYPES
*/
/*
	TODO: not sure how to struct this natively, but:
		  Dbus Struct:  (<symbol(s)...>) // Note: structs can be nested e.g. (i(ii))
		  Dbus Array:   a<symbol>		 // The symbol can be any type (even nested arrays, e.g. aai), but only one type is allowed. Arrays are like Golang slices; no fixed size.
		  Dbus Variant: v<symbol>		 // Dbus equivalent of interface{}, more or less. See https://dbus.freedesktop.org/doc/dbus-specification.html#container-types
		  Dbus Dict:    [kv]			 // Where k is the key's type and v is the value's type.
*/
