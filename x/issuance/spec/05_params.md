<!--
order: 5
-->

# Parameters

The issuance module has the following parameters:

| Key        | Type           | Example         | Description                                 |
|------------|----------------|-----------------|---------------------------------------------|
| Assets     | array (Asset)  | `[{see below}]` | array of assets created via issuance module |


Each `Asset` has the following parameters

| Key               | Type                   | Example                                         | Description                                           |
|-------------------|------------------------|-------------------------------------------------|-------------------------------------------------------|
| Owner             | sdk.AccAddress         | "nst1cd8z53n7gh2hvz0lmmkzxkysfp5pghufpe9j0r"   | the address that controls the issuance of the asset   |
| Denom             | string                 | "usdtoken"                                      | the denomination or exchange symbol of the asset      |
| BlockedAccounts   | array (sdk.AccAddress) | ["nst1tp9u8t8ang53a8tjh2mhqvvwdngqzjvmar0cpx"] | addresses which are blocked from holding the asset    |
| Paused            | boolean                | false                                           | boolean for if issuance and redemption are paused     |
