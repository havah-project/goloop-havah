# ChainScore APIs

## Introduction

This document is intended to explain the following:

* Basic and HAVAH-specific JSON-RPC APIs that ChainScore provides
* HAVAH-specific characteristics such as default configurations and predefined contracts

## Common

API path : `<scheme>://<host>/api/v3`

* All APIs follow SCORE API call convention
* Target SCORE Address for ChainScore APIs: `cx0000000000000000000000000000000000000000`
* For more details on SCORE API call convention, refer to [jsonrpc_v3.md](jsonrpc_v3.md)

## Basic APIs

Basic JSON-RPC APIs that ChainScore provides commonly, regardless of a specific platform

### setRevision

### getRevision

### setStepCost

### getStepCost

### getStepCosts

### setMaxStepLimit

### getMaxStepLimit

### setStepPrice

### getStepPrice

### getServiceConfig

### setScoreOwner

### getScoreOwner

### setRoundLimitFactor

### getRoundLimitFactor

### addDeployer

### removeDeployer

### isDeployer

### getDeployers

### setTimestampThreshold

### getTimestampThreshold

### grantValidator

### revokeValidator

### getValidators

## HAVAH APIs

HAVAH-specific JSON-RPC APIs

### startRewardIssue

* Set up the block height when to issue rewards begins
* Called by Governance SCORE

> Request
 
```json
{
  "jsonrpc": "2.0",
  "id": 1234,
  "method": "icx_sendTransaction",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "startRewardIssue",
      "params": {
        "height": "0x64"
      }
    }
  }
}
```

#### Parameters

| Key    | VALUE Type | Required | Description                              |
|:-------|:-----------|:---------|:-----------------------------------------|
| height | T_INT      | true     | Block height when to issue reward begins |

#### Returns

txHash

### addPlanetManager

* Adds a specified address to PlanetManager list
* Called by Governance SCORE

> Request
 
```json
{
  "jsonrpc": "2.0",
  "id": 1234,
  "method": "icx_sendTransaction",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "addPlanetManager",
      "params": {
        "address": "hx0123456789012345678901234567890123456789"
      }
    }
  }
}
```

#### Parameters

| Key     | VALUE Type | Required | Description           |
|:--------|:-----------|:---------|:----------------------|
| address | T_ADDRESS  | true     | PlanetManager address |

#### Returns

txHash

### removePlanetManager

* Removes a specific address from PlanetManager list
* Called by Governance SCORE
 
> Request
 
```json
{
  "jsonrpc": "2.0",
  "id": 1234,
  "method": "icx_sendTransaction",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "removePlanetManager",
      "params": {
        "address": "hx0123456789012345678901234567890123456789"
      }
    }
  }
}
```

#### Parameters

| Key     | VALUE Type | Required | Description           |
|:--------|:-----------|:---------|:----------------------|
| address | T_ADDRESS  | true     | PlanetManager address |

#### Returns

txHash

### isPlanetManager

* Query if a specific address is a PlanetManager or not

> Request
 
```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "icx_call",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "isPlanetManager",
      "params": {
        "address": "hx8f21e5c54f016b6a5d5fe65486908592151a7c57"
      }
    }
  }
}
```

> Response
 
```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "result": "0x1"
}
```

#### Parameters

| Key     | VALUE Type | Required | Description           |
|:--------|:-----------|:---------|:----------------------|
| address | T_ADDRESS  | true     | PlanetManager address |

#### Returns

| Key | VALUE Type | Required | Description           |
|:----|:-----------|:---------|:----------------------|
| -   | T_BOOL     | true     | true(0x0), false(0x1) |

### registerPlanet

* Registers a planet to the network
* Called by PlanetNFT SCORE

> Request
 
```json
{
  "jsonrpc": "2.0",
  "id": 1234,
  "method": "icx_sendTransaction",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "registerPlanet",
      "params": {
        "id": "0x1",
        "isPrivate": "0x0",
        "isCompany": "0x0",
        "owner": "hx8f21e5c54f016b6a5d5fe65486908592151a7c57",
        "usdt": "0x12a05f200", 
        "price": "0xa968163f0a57b400000"
      }
    }
  }
}
```

#### Parameters

| Key       | VALUE Type | Required | Description                             |
|:----------|:-----------|:---------|:----------------------------------------|
| id        | T_INT      | true     | Planet ID                               |
| isPrivate | T_BOOL     | true     | Is a private planet or not              |
| isCompany | T_BOOL     | true     | Is a company planet or not              |
| owner     | T_ADDRESS  | true     | Planet owner                            |
| usdt      | T_INT      | true     | Planet price in USDT (decimal: 10 ** 6) |
| price     | T_INT      | true     | Planet price in HVH (decimal: 10 ** 18) |

#### Returns

txHash

### getPlanetInfo

* Returns the information on the planet specified by id

> Request
 
```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "icx_call",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "getPlanetInfo",
      "params": {
        "id": "0x1"
      }
    }
  }
}
```

> Response

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "result": {
    "isCompany": "0x0",
    "isPrivate": "0x0",
    "owner": "hx0123456789012345678901234567890123456789",
    "usdtPrice": "0x12a05f200",
    "havahPrice": "0xa968163f0a57b400000",
    "height": "0x29"
  }
}
```

#### Parameters

| Key       | VALUE Type | Required | Description |
|:----------|:-----------|:---------|:------------|
| id        | T_INT      | true     | Planet ID   |

#### Returns

| Key        | VALUE Type | Required | Description                                |
|:-----------|:-----------|:---------|:-------------------------------------------|
| isPrivate  | T_BOOL     | true     | private flag                               |
| isCompany  | T_BOOL     | true     | company flag                               |
| owner      | T_ADDRESS  | true     | Planet owner                               |
| usdtPrice  | T_INT      | true     | Planet price in USDT                       |
| havahPrice | T_INT      | true     | Planet price in HVH                        |
| height     | T_INT      | true     | BlockHeight when the planet was registered |

### unregisterPlanet

* Unregisters a planet
* Called by PlanetNFT SCORE

> Request
 
```json
{
  "jsonrpc": "2.0",
  "id": 1234,
  "method": "icx_sendTransaction",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "unregisterPlanet",
      "params": {
        "id": "0x1"
      }
    }
  }
}
```

#### Parameters

| Key | VALUE Type | Required | Description |
|:----|:-----------|:---------|:------------|
| id  | T_INT      | true     | Planet ID   |

#### Returns

txHash

### setPlanetOwner

* Changes a planet owner
* Called by PlanetNFT SCORE

#### Parameters

| Key   | VALUE Type | Required | Description      |
|:------|:-----------|:---------|:-----------------|
| id    | T_INT      | true     | Planet ID        |
| owner | T_ADDRESS  | true     | New planet owner |

#### Returns

txHash

### reportPlanetWork

* PlanetManager reports a planet's work
* The network offers the rewards to a planet whose work has been reported in a term
* Only one report for a planet is available in a term
* `RewardOffered` eventlog is recorded in this transaction result
* Called by PlanetManager

> Request
 
```json
{
  "id": 1234,
  "jsonrpc": "2.0",
  "method": "icx_sendTransaction",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "reportPlanetWork",
      "params": {
        "id": "0x1"
      }
    }
  }
}
```

#### Parameters

| Key | VALUE Type | Required | Description      |
|:----|:-----------|:---------|:-----------------|
| id  | T_INT      | true     | Planet ID        |

#### Returns

txHash

#### EventLog

`RewardOffered(termSequence int, id int, rewardWithHoover int, hooverRequest int)`

| Key              | VALUE Type | Indexed | Description                                                              |
|:-----------------|:-----------|:--------|:-------------------------------------------------------------------------|
| termSequence     | T_INT      | false   | Term sequence starting with 0                                            |
| id               | T_INT      | false   | Planet ID                                                                |
| rewardWithHoover | T_INT      | false   | Rewards in HVH that the planet gets including an subsidy from HooverFund |
| hooverRequest    | T_INT      | false   | Subsidy from HooverFund                                                  |


### claimPlanetReward

* Claims remaining rewards for a specific planet
* Claimed rewards are transferred from PublicTreasury to the planet owner
* Called by a planet owner
 
> Request
 
```json
{
  "jsonrpc": "2.0",
  "id": 1234,
  "method": "icx_sendTransaction",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "claimPlanetReward",
      "params": {
        "ids": ["0x1", "0x2", "0x10"]
      }
    }
  }
}
```

#### Parameters

| Key | VALUE Type | Required | Description                 |
|:----|:-----------|:---------|:----------------------------|
| ids | T_LIST     | true     | Planet IDs to claim rewards |

#### Returns

txHash

### getRewardInfo

* Returns the information on a planet reward

> Request

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "icx_call",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "getRewardInfo",
      "params": {
        "id": "0x1"
      }
    }
  }
}
```

> Response
 
 ```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "result": {
    "height": "0x3e8",
    "total": "0x8ac7230489e80000",
    "remain": "0xde0b6b3a7640000",
    "claimable": "0xde0b6b3a7640000"
  }
}
```

#### Parameters

| Key | VALUE Type | Required | Description |
|:----|:-----------|:---------|:------------|
| id  | T_INT      | true     | Planet ID   |

#### Returns

| Key       | VALUE Type | Required | Description                                               |
|:----------|:-----------|:---------|:----------------------------------------------------------|
| height    | T_INT      | true     | Current block height                                      |
| total     | T_INT      | true     | Total accumulated rewards until now                       |
| remain    | T_INT      | true     | Difference between Total Rewards and Claimed Rewards      |
| claimable | T_INT      | true     | Rewards that a planet owner can receive when claiming now |

### getIssueInfo

* Returns the information on issue-related configuration
 
> Request
 
```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "icx_call",
  "params": {
    "from": "hxbe258ceb872e08851f1f59694dac2558708ece11",
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "getIssueInfo"
    }
  }
}
```

> Response

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "result": {
    "height": "0x41",
    "issueReductionCycle": "0x168",
    "issueStart": "0x14",
    "termPeriod": "0x1e",
    "termSequence": "0x1"
  }
}
```

#### Parameters

None
 
#### Returns

| Key                 | VALUE Type | Required | Description                                        |
|:--------------------|:-----------|:---------|:---------------------------------------------------|
| height              | T_INT      | true     | Current block height                               |
| termPeriod          | T_INT      | true     | Coin issuing term period (unit: block)             |
| issueReductionCycle | T_INT      | true     | issueAmount is reduced at a fixed rate every cycle |
| issueStart          | T_INT      | false    | BlockHeight when issuing coin will begin           |
| termSequence        | T_INT      | false    | Sequence of a term starting with 0                 |

### getUSDTPrice

* Returns 1 USDT price in HVH

> Request

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "icx_call",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "getUSDTPrice"
    }
  }
}
```

> Response
 
 ```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "result": "0x8ac7230489e80000"
}
```

#### Parameters

None

#### Returns

| Key | VALUE Type | Required | Description         |
|:----|:-----------|:---------|:--------------------|
| -   | T_INT      | true     | 1 USDT price in HVH |

### setUSDTPrice

* Set 1 USDT price in HVH
* Temporary API

> Request

```json
{
  "id": 1234,
  "jsonrpc": "2.0",
  "method": "icx_sendTransaction",
  "params": {
    "to": "cx0000000000000000000000000000000000000000",
    "dataType": "call",
    "data": {
      "method": "setUSDTPrice",
      "params": {
        "price": "0x8ac7230489e80000"
      }
    }
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description         |
|:------|:-----------|:---------|:--------------------|
| price | T_INT      | true     | 1 USDT price in HVH |

#### Returns

txHash
