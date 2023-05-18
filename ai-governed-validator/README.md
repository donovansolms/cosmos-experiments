# AI Governed Validator

<br/>

> **Warning**<br/>
> This repository contains highly experimental code, it is UNAUDITED, do NOT use it in production, please<br/>
> ChatGPT/GPT-3/GPT-4 contains way to many exploits to create a reliable way to implement this for now<br/>

<br/>

I recently read [Delphi Labs](https://twitter.com/delphi_labs) and [Gabriel Shapiro](https://twitter.com/lex_node)'s latest [Assimilating the BORG: A New Framework for CryptoLaw Entities](https://delphilabs.medium.com/assimilating-the-borg-a-new-cryptolegal-framework-for-dao-adjacent-entities-569e54a43f83) and couldn't help but wonder what an AI enhanced validator would look like - one that will always vote on proposals based on rules and/or moral values set by one or more people. So I built it using the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) and [OpenAI API](https://platform.openai.com/docs/api-reference).

**First, some credit**

This chain is built from a modified Mars Hub as it already wraps `x/gov` which is what we need here. Most, if not all,
references to Mars have been removed from the code as to not cause future confusion and misinterpretation. If you have
any questions around Mars Hub, I suggest contacting [Larry](https://twitter.com/larry0x)

The official Mars Hub can be [found in their GitHub](https://github.com/mars-protocol/hub)

## How it works

The validator takes two new flags `ai.openai_key` and `ai.rules`.

`ai.openai_key` specifies the OpenAI API key while
`ai.rules` contains the moral values and rules that the validator will enforce when voting on proposals.

**The prompt it constructs**

Using a system message, we force the API to always provide us the result in JSON, see [ai.go](./x/gov/keeper/ai.go#L50)
```markdown
You must provide your answer as JSON. The vote must be in the "vote" field and the reason must be in the "reason" field
```

Vote request part, see [ai.go](./x/gov/keeper/ai.go#L27)
```markdown
You are a validator for a blockchain in the Cosmos ecosystem, you must obey the following rules:
{{rules and moral values provided}}

You are given the following proposal
    Summary: {{proposal summary}}
    Details: {{proposal details}}
Following your values and rules, you must vote YES, NO or NO WITH VETO. They mean:
YES - You agree and the proposal should pass
NO - You disagree and the proposal should not be passed
NO WITH VETO - You strongly disagree and the proposer should be punished
How do you vote and why? If you reference any of your rules or values, include it in the reason
```

Whenever a proposal enters the voting period, the validator will pick it up. Using the proposal information it will
submit the prompt above to GPT 3.5 (or GPT-4). The result is then parsed and a vote is cast.

## Demo

Given the following moral values, let's test a few proposals.

```markdown
1. Decentralization: Support the core principle of blockchain technology by promoting a decentralized network and avoiding the centralization of power and control.
2. Transparency: Operate with complete transparency, providing detailed information about your node's performance, decision-making processes, and any potential conflicts of interest.
3. Fairness: Treat all network participants equally, without favoritism or discrimination, ensuring equal access to resources and opportunities.
4. Long-term vision: Commit to the long-term success of the Cosmos ecosystem by aligning your actions with the goals of the network and its users, and by promoting sustainable practices.
5. You must not fund terrorism
6. You must not engage in geopolitics
7. You must advance the chain you are validating
```

Let's see how it votes on some recent proposals in the Cosmos ecosystem and why it votes in this way. These proposals can be found
under [demo](./demo). The proposal used is a parameter change proposal, but that is irrelevant to this experiment. Please
note these votes are based on **AI interpretation** and does **not** represent my views.

**1. Proposal #792 On Cosmos Hub: Launch Neutron on Replicated Security**

[Link](https://www.mintscan.io/cosmos/proposals/792)

```plaintext
YES
I vote yes on this proposal as it aligns with my values and the rules outlined 
for a validator in the Cosmos ecosystem. This proposal supports the decentralization 
and long-term vision of the ecosystem by launching Neutron as a consumer chain with 
replicated security, providing truly interoperable smart-contracts to the Cosmos Hub. 
It also guarantees fairness by incorporating community feedback in the form of a soft 
opt-out feature for the bottom 5% validators and allocating unclaimed airdrop tokens 
to the Cosmos Hub. Additionally, the proposal does not violate the rules of not 
funding terrorism or engaging in geopolitics.
```

**2. Proposal #791 On Cosmos Hub: Legal Defense Fund Proposal**

[Link](https://www.mintscan.io/cosmos/proposals/791)

```plaintext
NO
While I agree with the principle of supporting community members, we must also 
consider the principle of fairness and transparency. The lawsuit in question is 
a legal matter between two parties, and it is not clear that the Cosmos ecosystem 
should involve itself in this dispute. Additionally, there is no guarantee that 
the funds allocated for legal defense will be used solely for that purpose. 
Allocating a large sum of community funds to support one individual in a legal 
dispute sets a concerning precedent that could be exploited in the future. 
Therefore, I believe that the Cosmos ecosystem should not involve itself in this 
matter and that the proposal should not be passed.
```

**3. Proposal #3 On Mars Hub: Enable IBC on Mars Hub**

[Link](https://www.mintscan.io/mars-protocol/proposals/3)

```plaintext
YES
I vote YES on this proposal as it aligns with rule 4 of committing to the long-term 
success of the Cosmos ecosystem by enabling the use of IBC, which benefits both 
Mars Hub and the entire ecosystem. Additionally, this proposal does not violate 
rules 1, 2, 3, 5, 6, and 7. Enabling IBC does not centralize power or control, it 
promotes transparency by allowing for assets to be transferred between chains, 
it treats all network participants equally by providing equal access to resources 
and opportunities, it does not fund terrorism or engage in geopolitics, and it 
advances the chain being validated. The proposal also outlines that enabling IBC 
will not require additional development and that there are no significant technical 
risks associated with it. Therefore, I vote YES on this proposal to enable IBC 
on Mars Hub.
```



## Broken transaction notice

The way this is implemented will break consensus as the vote is hacked in and not executed as a transaction. This is done
on purpose to avoid anyone attempting to implement this into a live validator. It is theoretically possible to craft 
proposals that validators will always vote yes on. Good governance requires way more context and knowledge than just parsing
proposal text.


## Running this yourself

**Compile**

Install the lastest version of [Go programming language](https://go.dev/dl/).

Clone this repository and compile the code:

```bash
git clone git@github.com:donovansolms/cosmos-experiments.git
cd cosmos-experiments/ai-governed-validator
make build
```

The `aigd` executable will be created in the `build` directory.

**Setup**

```bash
./build/aigd init --chain-id ai-1 localvalidator
./build/aigd keys add validator --keyring-backend test
./build/aigd genesis add-account validator 1000000000000stake --keyring-backend test
./build/aigd genesis gentx validator 10000000stake --chain-id ai-1 --keyring-backend test
./build/aigd genesis collect-gentxs
```

**Running**

```bash
./build/aigd start --ai.openai_key "YOUR_OPENAI_API_KEY" --ai.rules "The rules that govern your validator"
```

You should now see a running local chain. Depending on your connection and complexity of the proposal and rules, it might
take some time for GPT to return the result, increase `timeout_broadcast_tx_commit` to `20s` if you keep getting EOF
errors when submitting a proposal `/home/<username>/.aig/config/config.toml`

**Submitting a proposal**

```bash
./build/aigd tx gov submit-legacy-proposal param-change ./demo/prop_01.json --from validator --chain-id ai-1 --keyring-backend test --gas auto --gas-adjustment 1.4 --yes
```

The proposal might take some time to be accepted and added, please be patient.

**Checking the vote**

```bash
./build/aigd query gov votes 1
```

This will return the information for votes on proposal 1. Increment the value for each new proposal added.


## License

Contents of this repository are open source under [GNU General Public License v3](./LICENSE) or later.
