package gosecret

// This currently is not used.

/*
	TranslateError translates a SecretServiceErrEnum into a SecretServiceError.
	If a matching error was found, ok will be true and err will be the matching SecretServiceError.
	If no matching error was found, however, then ok will be false and err will be ErrUnknownSecretServiceErr.
*/
func TranslateError(ssErr SecretServiceErrEnum) (ok bool, err error) {

	err = ErrUnknownSecretServiceErr

	for _, e := range AllSecretServiceErrs {
		if e.ErrCode == ssErr {
			ok = true
			err = e
			return
		}
	}

	return
}

// Error returns the string format of the error; this is necessary to be considered a valid error interface.
func (e SecretServiceError) Error() (errStr string) {

	errStr = e.ErrDesc

	return
}
