# Genesis Transaction for HAVAH

## Introduction

* This document explains the havah-specific configurations in genesis file format
* Common configuration is also available in havah-specific genesis file
* For more details on common configuration, refer to [genesis_tx.md](genesis_tx.md)
 
## Predefined addresses

| Name            | Address                                      |
|:----------------|:---------------------------------------------|
| ChainScore      | `cx0000000000000000000000000000000000000000` |
| Governance      | `cx0000000000000000000000000000000000000001` |
| SystemTreasury  | `hx1000000000000000000000000000000000000000` |
| PublicTreasury  | `hx3000000000000000000000000000000000000000` |
| SustainableFund | `cx4000000000000000000000000000000000000000` |
| HooverFund      | `hx6000000000000000000000000000000000000000` |
| EcoSystem       | `cx7000000000000000000000000000000000000000` |
| PlanetNFT       | `cx8000000000000000000000000000000000000000` |
| ServiceTreasury | `hx9000000000000000000000000000000000000000` |
| Vault           | `cx1100000000000000000000000000000000000000` |

## HAVAH-specific Configuration

### Example

```json
{
  ...
  "chain": {
    ...
    "platform": "havah",
    "havah": {
      "termPeriod": "0xa8c0",
      "issueReductionCycle": "0x168",
      "issueLimit": "0x4650",
      "issueAmount": "0x38e8f7792d79767800000",
      "hooverBudget": "0x38e8f7792d79767800000",
      "usdtPrice": "0x8ac7230489e80000"
    }
  }
}
```

### Parameters

| Key                 | VALUE type | Required |              Default |  Unit |
|:--------------------|:-----------|:---------|---------------------:|------:|
| termPeriod          | T_INT      | false    |                43200 | block |
| issueReductionCycle | T_INT      | false    |                  360 |  term |
| issueLimit          | T_INT      | false    |                18000 |  term |
| issueAmount         | T_INT      | false    | 4_300_000 * 10 ** 18 |   HVH |
| hooverBudget        | T_INT      | false    | 4_300_000 * 10 ** 18 |   HVH |
| usdtPrice           | T_INT      | true     |                    - |   HVH |

* `termPeriod`: Coins for reward are issued every term period in blocks
* `issueReductionCycle`: issueAmount is reduced at a fixed rate each cycle
* `issueLimit`: No issues after issueLimit
* `issueAmount`: Amount of coins to be issued every term period
* `hooverBudget`: Max budget of HooverFund account  
* `usdtPrice`: 1 USDT price in HVH
 