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
      "privateReleaseCycle": "0x1e",
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

| Key                 | VALUE type | Possible to omit |              Default | Description                               |
|:--------------------|:-----------|:-----------------|---------------------:|-------------------------------------------|
| termPeriod          | T_INT      | true             |                43200 | term period in blocks when coin is issued |
| issueReductionCycle | T_INT      | true             |                  360 | unit: term                                |
| privateReleaseCycle | T_INT      | true             |                   30 | unit: term                                |
| privateLockup       | T_INT      | true             |                  360 | unit: term                                |
| issueLimit          | T_INT      | true             |                18000 | No issues after issueLimit (unit: term)   |
| issueAmount         | T_INT      | true             | 4_300_000 * 10 ** 18 | unit: HVH                                 |
| hooverBudget        | T_INT      | true             | 4_300_000 * 10 ** 18 | unit: HVH                                 |                                         |
| usdtPrice           | T_INT      | false            |                    - | 1 USDT price in HVH                       |
