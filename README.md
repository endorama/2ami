# two-factor-authenticator agent

Two factor authenticator agent that stores 2FA secrets in system keyring, thus
avoiding having them in cleartext (and accessible) somewhere on your computer.

**NB: Current version is beta state**

## Keyring/Keychain encryption

OTP Secret keys are directly saved and retrieved from system keyring, and are not
being kept in process active memory.

Supported secret storage backends are:
- macOS/OSX Keychain
- Secret Service ( Gnome )

A full list of of the available backends can be found [here](https://github.com/99designs/keyring).

*Security considerations:* the secrets are still beign loaded in memory when
adding a new key and generating a new token, even if for a small amount of time.
I believe this is the safest approach, please correct me if I'm wrong. :)

**Note:** This software is in beta and has not been security reviewed (yet).

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

- as per [99designs/keyring#22](https://github.com/99designs/keyring/pull/22)
  currently the `KDE Wallet` backend cannot be enabled
