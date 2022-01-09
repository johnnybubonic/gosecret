// See LICENSE in source root directory for copyright and licensing information.

/*
Package gosecret is(/was originally) a fork of go-libsecret (see https://github.com/gsterjov/go-libsecret
and https://pkg.go.dev/github.com/gsterjov/go-libsecret).

It was forked in order to present bugfixes, actually document the library, conform to more Go-like patterns, and
provide missing functionality (as the original seems to be unmaintained).
As such, hopefully this library should serve as a more effective libsecret/SecretService interface.

Backwards Compatibility

Version series `v0.X.X` of this library promises full and non-breaking backwards compatibility/drop-in
support of API interaction with the original project.
The only changes should be internal optimizations, adding documentation, some file reorganizing, adding Golang module support,
etc. -- all transparent from the library API itself.

To use this library as a replacement without significantly modifying your code,
you can simply use a `replace` directive in your go.mod file:

	// ...
	replace (
		github.com/gsterjov/go-libsecret dev => r00t2.io/gosecret v0
	)

and then run `go mod tidy`.

Do NOT use the master branch. For anything. I make no promises on the stability of that branch at any given time.
New features will be added to V1 branch, and stable releases will be tagged. V0 branch is reserved only for optimization and bug fixes.

New Developer API

Starting from `v1.0.0` onwards, entirely breaking changes can be assumed from the original project.
To use the new version,

	import (
		`r00t2.io/gosecret/v1`
	)

To reflect the absolute breaking changes, the module name changes as well from `libsecret` to `gosecret`.

SecretService Concepts

For reference:

- A Service allows one to retrieve and operate on/with Session and Collection objects.

- A Session allows one to operate on/with Item objects (e.g. parsing/decoding/decrypting them).

- A Collection allows one to retrieve and operate on/with Item objects.

- An Item allows one to retrieve and operate on/with Secret objects.

(Secrets are considered "terminating objects" in this model, and contain actual secret value(s) and metadata).

Various interactions are handled by Prompts.

So the object hierarchy in THEORY looks kind of like this:

	Service
	├─ Session "A"
	├─ Session "B"
	├─ Collection "A"
	│	├─ Item "A.1"
	│	│	├─ Secret "A_1_a"
	│	│	└─ Secret "A_1_b"
	│	└─ Item "A.2"
	│		├─ Secret "A_2_a"
	│		└─ Secret "A_2_b"
	└─ Collection "B"
		├─ Item "B.1"
		│	├─ Secret "B_1_a"
		│	└─ Secret "B_1_b"
		└─ Item "B.2"
			├─ Secret "B_2_a"
			└─ Secret "B_2_b"

And so on.
In PRACTICE, however, most users will only have two Collection items
(a default "system" one named "login", which usually is unlocked upon login,
and a temporary one that may or may not exist, running in memory for the current login session named `session`).

Usage

Full documentation can be found via inline documentation.
Additionally, use either https://pkg.go.dev/r00t2.io/gosecret or https://pkg.go.dev/golang.org/x/tools/cmd/godoc (or `go doc`) in the source root.

Note that many functions/methods may return a (r00t2.io/goutils/)multierr.MultiError (https://pkg.go.dev/r00t2.io/goutils/multierr#MultiError),
which you may attempt to typeswitch back to a *multierr.MultiErr to receive the original errors in their native error format (MultiError.Errors).
The functions/methods which may return a MultiError are noted as such in their individual documentation.
*/
package gosecret
