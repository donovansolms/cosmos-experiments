# AI Governed Chain

"A DAO must encode all its rules in its software" stood out to me when I read Delphi Labs and Gabriel Shapiro's latest [Assimilating the BORG: A New Framework for CryptoLaw Entities](https://delphilabs.medium.com/assimilating-the-borg-a-new-cryptolegal-framework-for-dao-adjacent-entities-569e54a43f83) I couldn't help but wonder if AI could be added to validators to 


## Credit

This chain is built from a modified Mars Hub as it already wraps x/gov. Most, if not all,
references to Mars has been removed from the code as to not cause future confusion. 

Mars Hub app-chain, built on top of [Tendermint][1], [Cosmos SDK][2], [IBC][3], and [CosmWasm][4].

## Installation

Install the lastest version of [Go programming language][5] and configure related environment variables. See [here][6] for a tutorial.

Clone this repository, checkout to the latest tag, the compile the code:

```bash
git clone https://github.com/mars-protocol/hub.git
cd hub
git checkout <tag>
make install
```

A `aigd` executable will be created in the `$GOBIN` directory.

## License

Contents of this repository are open source under [GNU General Public License v3](./LICENSE) or later.

[1]: https://github.com/tendermint/tendermint
[2]: https://github.com/cosmos/cosmos-sdk
[3]: https://github.com/cosmos/ibc-go
[4]: https://github.com/CosmWasm/wasmd
[5]: https://go.dev/dl/
[6]: https://github.com/larry0x/workshops/tree/main/how-to-run-a-validator
