package gosecret

import (
	"errors"
)

// General errors.
var (
	// ErrBadDbusPath indicates an invalid path - either nothing exists at that path or the path is malformed.
	ErrBadDbusPath error = errors.New("invalid dbus path")
	// ErrInvalidProperty indicates a dbus.Variant is not the "real" type expected.
	ErrInvalidProperty error = errors.New("invalid variant type; cannot convert")
	// ErrNoDbusConn gets triggered if a connection to Dbus can't be detected.
	ErrNoDbusConn error = errors.New("no valid dbus connection")
)

/*
	Translated SecretService errors.
	See https://developer-old.gnome.org/libsecret/unstable/libsecret-SecretError.html#SecretError.
	Used by TranslateError.
*/
var (
	ErrUnknownSecretServiceErr error              = errors.New("cannot find matching SecretService error")
	ErrSecretServiceProto      SecretServiceError = SecretServiceError{
		ErrCode: EnumErrProtocol,
		ErrName: "SECRET_ERROR_PROTOCOL",
		ErrDesc: "an invalid message or data was received from SecretService",
	}
	ErrSecretServiceLocked SecretServiceError = SecretServiceError{
		ErrCode: EnumErrIsLocked,
		ErrName: "SECRET_ERROR_IS_LOCKED",
		ErrDesc: "the item/collection is locked; the specified operation cannot be performed",
	}
	ErrSecretServiceNoObj SecretServiceError = SecretServiceError{
		ErrCode: EnumErrNoSuchObject,
		ErrName: "SECRET_ERROR_NO_SUCH_OBJECT",
		ErrDesc: "no such item/collection was found in SecretService",
	}
	ErrSecretServiceExists SecretServiceError = SecretServiceError{
		ErrCode: EnumErrAlreadyExists,
		ErrName: "SECRET_ERROR_ALREADY_EXISTS",
		ErrDesc: "a relevant item/collection already exists",
	}
	ErrSecretServiceInvalidFormat SecretServiceError = SecretServiceError{
		ErrCode: EnumErrInvalidFileFormat,
		ErrName: "SECRET_ERROR_INVALID_FILE_FORMAT",
		ErrDesc: "the file/content format is invalid",
	}
	/*
		AllSecretServiceErrs provides a slice of these for easier iteration when translating.
		TECHNICALLY, because they are indexed in the order of their enums, you could
		simplify and optimize translation by just doing e.g.

			err = AllSecretServiceErrs[EnumErrProtocol]

		But this should be considered UNSTABLE and UNSAFE due to it being potentially unpredictable in the future.
		There are only 5 errors currently, so the performance benefits would be negligible compared to iteration.
		If SecretService adds more errors, however, this may be more desirable.
	*/
	AllSecretServiceErrs []SecretServiceError = []SecretServiceError{
		ErrSecretServiceProto,
		ErrSecretServiceLocked,
		ErrSecretServiceNoObj,
		ErrSecretServiceExists,
		ErrSecretServiceInvalidFormat,
	}
)
