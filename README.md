# two-factor-authenticator agent

Two factor authenticator agent that stores 2FA secrets in system keyring, thus
avoiding having them in cleartext (and accessible) somewhere on your computer.

## Keyring/Keychain encryption

OTP Secret keys are directly saved and retrieved from system keyring, and are not
being kept in process active memory.

Enabled secret storage backends are:
- macOS/OSX Keychain
- Secret Service ( Gnome )

but more are available.

A full list of of the available backends can be found [here](https://github.com/99designs/keyring). If you are interested and able to test with the specified backend, just open a issue and I'll be happy to have a look.

*Security considerations:* the secrets are still being loaded in memory when adding a new key and generating a new token, even if for a small amount of time.
I believe this is a safe enough approach ( as in for general use case ), and is surely better than plain secrets on file system. 
Please correct me if I'm wrong. :)

**Note:** This software has **not** been security reviewed by a third party.

## Installation

Go to the [Release tab](https://github.com/endorama/two-factor-authenticator/releases) and grab your executable. Download it and add execution permissions.

You can watch for new releases through GitHub by watching the repository!

## Usage

```
Usage:
  two-factor-authenticator add <name> [--digits=<digits>] [--interval=<seconds>] [--verbose]
  two-factor-authenticator generate <name> [-c|--clip] [--verbose]
  two-factor-authenticator list [--verbose]
  two-factor-authenticator remove <name> [--verbose]
  two-factor-authenticator -h | --help
  two-factor-authenticator --version

Commands:
  add       Add a new key.
  generate  Generate a token from a known key.
  list      List known keys.
  remove    Remove specified key.

Options:
  -h --help             Show this screen.
  --version             Show version.
  --verbose             Enable verbose output.
  --digits=<digits>     Number of token digits
  --interval=<seconds>  Interval in seconds between token generation
  -c --clip             Copy result to the clipboard
```

## Generated tokens

Generated token are formatted as Google Authenticator: zeros are prepended in
place of missing digits.

## TODO

- custom token formatters
- backup/restore functionalities

## Known issues

None.
