# v3onion

v3 .onion vanity address generator

Inspired by [onion-go](https://github.com/rdkr/oniongen-go) and [bine](github.com/cretz/bine) I've decided to make my version of v3 vanity address generator 

### Installation
```bash
> git clone https://github.com/koceg/v3onion.git
> cd v3onion
> go build
```
## Usage

```bash
> ./v3onion -c 1 -r "^test" # number of matches and regex expression to use for the matching result
```

## Notes

This might not be the fastest generator available but I find it simple enough, and the termination mechanism might allow 1-2 extra (but not less) addresses that were in flight during thread termination. 

Your matches will be saved under *v3* directory in the current working path.

Afterwords copy the content of the desired key to the path where tor is going to search for the hidden service keys. Tor is going to generate the right address under hostname file automatically. The process where we calculate the address hash is skipped in this program because **onion_address = base32(pubkey || checksum || version)** is wasted CPU cycles and we USUALLY search for a few characters in the public key.