<h1 align="center">2ami</h1>
<p align="center">Your easy 2FA companion that keep the secrets secret.</p>

<p align="center">
  <img alt="Go report score" src="https://goreportcard.com/badge/github.com/endorama/2ami" >
  <img alt="GitHub Release Date" src="https://img.shields.io/github/release-date/endorama/2ami?color=blue">
  <img alt="GitHub tag (latest by date)" src="https://img.shields.io/github/v/tag/endorama/2ami?label=latest">
  <img alt="Share on Twitter" src="https://img.shields.io/twitter/url?style=social&url=https%3A%2F%2Fgithub.com%2Fendorama%2F2ami%2F" >
</p>

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

## Stargazers
[![Stargazers repo roster for @endorama/2ami](https://reporoster.com/stars/endorama/2ami)](https://github.com/endorama/2ami/stargazers)

## Contributors
[![Forkers repo roster for @endorama/2ami](https://reporoster.com/forks/endorama/2ami)](https://github.com/endorama/2ami/network/members)
