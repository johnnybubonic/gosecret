package gosecret

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

func (e SecretServiceError) Error() (errStr string) {

	errStr = e.ErrDesc

	return
}
