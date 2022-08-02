![go report](https://goreportcard.com/badge/github.com/endorama/2ami)

<img alt="2ami Secure Two Factor Authenticator logo" src="images/logo.svg" width="100%" />

`2ami` is a two factor authenticator for the CLI that stores 2FA secrets in system keyring, avoiding storing them in cleartext on your computer.

OTP Secret keys are saved and retrieved from system keyring at each use, so are not being kept in process active memory if not during operation explicitly requiring them.

**Security considerations:** the secrets are still being loaded in memory when adding a new key and generating a new token, even if for a small amount of time.
I believe this is a safe enough approach (in a normal threat model, please consider yours), and is surely better than plain secrets on file system. 
Happy to discuss security improvements! :)

**Note:** This software has **not** been security reviewed by a third party.

Interested in using it? [Look at the getting started page](https://github.com/endorama/2ami/wiki/Getting-Started).

What to dig deeper? Go to the [project wiki](https://github.com/endorama/2ami/wiki).


## Known issues

None.

## Contributors

[@backwards-rat-race](https://github.com/backwards-rat-race)
