package gosecret

import (
	`fmt`
)

/*
	MarshalJSON converts a SecretValue to a JSON representation.
	For compat reasons, the MarshalText is left "unmolested" (i.e. renders to a Base64 value).
	I don't bother with an UnmarshalJSON because it makes exactly 0 sense to unmarshal due to runtime and unexported fields in Secret.
*/
func (s *SecretValue) MarshalJSON() (b []byte, err error) {

	b = []byte(fmt.Sprintf("\"%v\"", string(*s)))

	return
}
