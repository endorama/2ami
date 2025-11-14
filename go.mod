module github.com/endorama/2ami

go 1.15

require (
	github.com/99designs/keyring v1.1.6
	github.com/OpenPeeDeeP/xdg v1.0.0
	github.com/atotto/clipboard v0.1.0
	github.com/boltdb/bolt v1.3.1
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/docopt/docopt.go v0.0.0-20180111231733-ee0de3bc6815
	github.com/dvsekhvalnov/jose2go v1.7.0 // indirect
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/hgfischer/go-otp v1.0.0
	github.com/keybase/go-keychain v0.0.0-20201121013009-976c83ec27a6 // indirect
	github.com/mitchellh/cli v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.9.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
)

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
