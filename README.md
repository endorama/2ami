![go report](https://goreportcard.com/badge/github.com/endorama/2ami)

<img alt="2ami Secure Two Factor Authenticator logo" src="images/logo.svg" width="100%" />

`2ami` is a two factor authenticator for the CLI that stores 2FA secrets in system keyring, avoiding storing them in cleartext on your computer.

OTP Secret keys are saved and retrieved from system keyring at each use, so are not being kept in process active memory if not during operation explicitly requiring them.

**Security considerations:** the secrets are still being loaded in memory when adding a new key and generating a new token, even if for a small amount of time.
I believe this is a safe enough approach (in a normal threat model, please consider yours), and is surely better than plain secrets on file system. 
Happy to discuss security improvements! :)

**Note:** This software has **not** been security reviewed by a third party.

## Keyring/Keychain encryption

Enabled secret storage backends are:
* [macOS Keychain](https://support.apple.com/en-au/guide/keychain-access/welcome/mac)
* [Windows Credential Manager](https://support.microsoft.com/en-au/help/4026814/windows-accessing-credential-manager)
* Secret Service ([Gnome Keyring](https://wiki.gnome.org/Projects/GnomeKeyring), [KWallet](https://kde.org/applications/system/org.kde.kwalletmanager5))

More storage are available, a full list can be found [here](https://github.com/99designs/keyring). If you are interested and able to test with the specified backend, just open a issue to have it added.

## Installation

Go to the [Release tab](https://github.com/endorama/2ami/releases) and grab your executable. Download it and add execution permissions.

You can watch for new releases through GitHub by watching the repository!

## Usage

```
$ 2ami
Two factor authenticator for your command line.

Usage:
  2ami add <name> [--digits=<digits>] [--interval=<seconds>] [--verbose]
  2ami dump [<name>] [--verbose]
  2ami generate <name> [-c|--clip] [--verbose]
  2ami list [--verbose]
  2ami remove <name> [--verbose]
  2ami rename <old-name> <new-name>
  2ami backup <file-path>
  2ami restore <file-path>
  2ami -h | --help
  2ami --version

Commands:
  add       Add a new key.
  dump      Dump keys informations (without secrets).
  generate  Generate a token from a known key.
  list      List known keys.
  remove    Remove specified key.
  backup    Backup keys to a specified file
  restore   Restore keys from a specified file

Options:
  -h --help             Show this screen.
  --version             Show version.
  --verbose             Enable verbose output.
  --digits=<digits>     Number of token digits.
  --interval=<seconds>  Interval in seconds between token generation.
  -c --clip             Copy result to the clipboard.

Environment variables:
  2AMI_DB    Path to the database where 2FA keys information are stored.
             Default to $XDG_DATA_HOME/2ami/database.boltdb.
						 For non Linux values of XDG_DATA_HOME see https://github.com/OpenPeeDeeP/xdg
	2AMI_RING	 Name of the keyring/keychain where 2FA secrets will be stored.
						 Default to "login".
```

## Generated tokens

Generated token are formatted as Google Authenticator: zeros are prepended in
place of missing digits.

Custom formatters may be implemented if needed.

## Known issues

None.

## Contributors

[@backwards-rat-race](https://github.com/backwards-rat-race)
