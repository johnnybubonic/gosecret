package gosecret

// NewSecret returns a pointer to a new Secret based on a Session, parameters, (likely an empty byte slice), a value, and the MIME content type.
func NewSecret(session *Session, params []byte, value []byte, contentType string) (secret *Secret) {

	secret = &Secret{
		Session:     session.Path(),
		Parameters:  params,
		Value:       value,
		ContentType: contentType,
	}

	return
}
