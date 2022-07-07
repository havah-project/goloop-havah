# Genesis Transaction for HAVAH

## Introduction

* This document explains the havah-specific configurations in genesis file format
* Common configuration is also available in havah-specific genesis file
* For more details on common configuration, refer to [genesis_tx.md](genesis_tx.md)

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
      "privateReleaseCycle": "0x18",
      "privateLockup": "0x168",
      "issueLimit": "0x4650",
      "issueAmount": "0x38e8f7792d79767800000",
      "hooverBudget": "0x38e8f7792d79767800000",
      "usdtPrice": "0x8ac7230489e80000"
    }
  }
}
```

### Parameters

| Key                 | VALUE type | Possible to omit |              Default |  Unit |
|:--------------------|:-----------|:-----------------|---------------------:|------:|
| termPeriod          | T_INT      | true             |                43200 | block |
| issueReductionCycle | T_INT      | true             |                  360 |  term |
| privateReleaseCycle | T_INT      | true             |                   30 |  term |
| privateLockup       | T_INT      | true             |                  360 |  term |
| issueLimit          | T_INT      | true             |                18000 |  term |
| issueAmount         | T_INT      | true             | 4_300_000 * 10 ** 18 |   HVH |
| hooverBudget        | T_INT      | true             | 4_300_000 * 10 ** 18 |   HVH |                                         |
| usdtPrice           | T_INT      | false            |                    - |   HVH |

* `termPeriod`: Coins for reward are issued every term period in blocks
* `issueReductionCycle`: issueAmount is reduced at a fixed rate each cycle
* `privateReleaseCycle`: Additional 1/24 of total rewards can be claimed each cycle
* `privateLockup`: Period during which the reward is locked up
* `issueLimit`: No issues after issueLimit
* `issueAmount`: Amount of coins to be issued every term period
* `hooverBudget`: Max budget of HooverFund account  
* `usdtPrice`: 1 USDT price in HVH
 