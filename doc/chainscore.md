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

### setRevision(code int)

* Sets a revision to activate new features
* Called by Governance SCORE
 
> Request

```json
{
  "data": {
    "method": "setRevision",
    "params": {
      "code": "0x13"
    }
  }
}
```

#### Parameters

| Key  | VALUE Type | Required | Description   |
|:-----|:-----------|:---------|:--------------|
| code | T_INT      | true     | revision code |

#### Returns

`T_HASH` - txHash

### getRevision() int

* Returns revision

> Request
 
 ```json
{
  "data": {
    "method": "getRevision"
  }
}
```

> Response

```json
{
  "result": "0x13"
}
```

#### Parameters

None

#### Returns

`T_INT` - Revision code

### setStepCost(type string, cost int)

* Sets step cost of each action
* Called by Governance SCORE

> Request

```json
{
  "data": {
    "method": "setStepCost",
    "params": {
      "type": "apiCall",
      "cost": "0x2710"
    }
  }
}
```

#### Parameters

| Key  | VALUE Type | Required | Description          |
|:-----|:-----------|:---------|:---------------------|
| type | T_STRING   | true     | action type          |
| cost | T_INT      | true     | step cost for action |

> Action types

| Action         |        Default | Description |
|:---------------|---------------:|:------------|
| default        |        100_000 | -           |
| contractCall   |         25_000 | -           |
| contractCreate |  1_000_000_000 | -           |
| contractUpdate |  1_000_000_000 | -           |
| contractSet    |         15_000 | -           |
| getBase        |          3_000 | -           |
| get            |             25 | -           |
| setBase        |         10_000 | -           |
| set            |            320 | -           |
| deleteBase     |            200 | -           |
| delete         |           -240 | -           |
| input          |            200 | -           |
| logBase        |          5_000 | -           |
| log            |            100 | -           |
| apiCall        |         10_000 | -           |

#### Returns

`T_HASH` - txHash

### getStepCost(type string) int

* Returns the step cost of a specific action

> Request

```json
{
  "data": {
    "method": "getStepCost",
    "params": {
      "type": "apiCall"
    }
  }
}
```

> Response
 
 ```json
{
  "result": "0x2710"
}
```

#### Parameters

| Key  | VALUE Type | Required | Description |
|:-----|:-----------|:---------|:------------|
| type | T_STRING   | true     | action type |

#### Returns

`T_INT` - step cost of a specific action type

### getStepCosts() dict

* Returns a table of the step costs for each action

> Request
 
```json
{
  "data": {
    "method": "getStepCosts"
  }
}
```

> Response
 
```json
{
  "result": {
    "schema": "0x1",
    "default": "0x186a0",
    "input": "0xc8",
    "contractCall": "0x61a8",
    "contractCreate": "0x3b9aca00",
    "contractUpdate": "0x3b9aca00",
    "contractSet": "0x3a98",
    "get": "0x19",
    "getBase": "0xbb8",
    "set": "0x140",
    "setBase": "0x2710",
    "delete": "-0xf0",
    "deleteBase": "0xc8",
    "log": "0x64",
    "logBase": "0x1388",
    "apiCall": "0x2710"
  }
}
``` 

#### Parameters

None

#### Returns

`T_DICT` - a dict: key - camel-cased action strings, value - step costs in integer

### setMaxStepLimit(contextType string, limit int)

* Sets the maximum step limit that any SCORE execution should be bounded by
* Called by Governance SCORE
 
> Request
 
```json
{
  "data": {
    "method": "setMaxStepLimit",
    "params": {
      "contextType": "query",
      "limit": "0x2faf080"
    }
  }
}
```

#### Parameters

| Key         | VALUE Type | Required | Description                                    |
|:------------|:-----------|:---------|:-----------------------------------------------|
| contextType | T_STRING   | true     | `invoke` for sendTransaction, `query` for call |
| limit       | T_INT      | true     | maximum step limit for each contextType        |

#### Returns

`T_HASH` - txHash

### getMaxStepLimit(contextType string) int

* Returns the maximum step limit value that any SCORE execution should be bounded by

> Request
 
```json
{
  "data": {
    "method": "getMaxStepLimit",
    "params": {
      "contextType": "invoke"
    }
  }
}
``` 

> Response

```json
{
  "result": "0x9502f900"
}
```

#### Parameters

| Key         | VALUE Type | Required | Description                                    |
|:------------|:-----------|:---------|:-----------------------------------------------|
| contextType | T_STRING   | true     | `invoke` for sendTransaction, `query` for call |

#### Returns

`T_INT` - integer of the maximum step limit for the given contextType

### setStepPrice(price int)

* Sets the new step price
* Called by Governance SCORE

> Request
 
```json
{
  "data": {
    "method": "setStepPrice",
    "params": {
      "price": "0x2e90edd00"
    }
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description    |
|:------|:-----------|:---------|:---------------|
| price | T_INT      | true     | new step price |

#### Returns

`T_HASH` - txHash

### getStepPrice() int

* Returns the current step price

> Request

```json
{
  "data": {
    "method": "getStepPrice"
  }
}
```

> Response
 
```json
{
  "result": "0x2e90edd00"
}
```

#### Parameters

None

#### Returns

`T_INT` - step price

### getServiceConfig() int

* Returns an integer value representing service configuration bitwise flags

> Request
 
```json
{
  "data": {
    "method": "getServiceConfig"
  }
}
``` 

> Response

```json
{
  "result": "0x1"
}  
```
  
#### Parameters

None

#### Returns

`T_INT` - integer value representing the service configuration bitwise flags

> Service configuration flags

| Name                  | VALUE | 
|:----------------------|:------|
| Fee                   | 0x01  |
| Audit                 | 0x02  |
| DeployerWhiteList     | 0x04  |
| ScorePackageValidator | 0x08  |
| Membership            | 0x10  |
| FeeSharing            | 0x20  |

### setScoreOwner(score Address, owner Address)

* Changes the owner of the score indicated by a given address
* Only the score owner can change its owner
* If a score owner changes its owner to `hx0000000000000000000000000000000000000000`, it means that the score is frozen and no one can update or modify it anymore
* Score address is also available for a score owner
* A score itself can be set to its owner
* Called by Governance SCORE

> Request

```json
{
  "data": {
    "method": "setScoreOwner",
    "params": {
      "score": "cx8d3ef83a63d8bbd3f08c4a8b8a18fbae13368b40",
      "owner": "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
    }
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description                        |
|:------|:-----------|:---------|:-----------------------------------|
| score | T_ADDRESS  | true     | score address to change its owner  |
| owner | T_ADDRESS  | true     | new owner address of a given score |

#### Returns

`T_HASH` - txHash

### getScoreOwner(score Address) Address

* Returns the owner of the score indicated by a given address

> Request

```json
{
  "data": {
    "method": "getScoreOwner",
    "params": {
      "score": "cx8d3ef83a63d8bbd3f08c4a8b8a18fbae13368b40"
    }
  }
}
```

> Response

```json
{
  "result": "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
}
```

#### Parameters

| Key   | VALUE Type | Required | Description            |
|:------|:-----------|:---------|:-----------------------|
| score | T_ADDRESS  | true     | score address to query |

#### Returns

`T_ADDRESS` - owner address of a given score

### setTimestampThreshold(threshold int)

* Sets transaction timestamp threshold in millisecond
* Transactions whose timestamp is out of range of timestamp threshold is rejected
* `block.timestamp - threshold <= valid tx timestamp < block.timestamp + threshold`
* Called by Governance SCORE

> Request
 
```json
{
  "data": {
    "method": "setTimestampThreshold",
    "params": {
      "threshold": "0x493e0"
    }
  }
}
```

#### Parameters

| Key       | VALUE Type | Required | Description        |
|:----------|:-----------|:---------|:-------------------|
| threshold | T_INT      | true     | tx threshold in ms |

#### Returns

`T_HASH` - txHash

### getTimestampThreshold() int

* Returns transaction threshold in millisecond

> Request
 
```json
{
  "data": {
    "method": "getTimestampThreshold"
  }
}
```

> Response
 
```json
{
  "result": "0x493e0"
}
```

#### Parameters

None

#### Returns
 
`T_INT` - transaction threshold in millisecond

### addDeployer(address Address)

* Adds an address to deployer list
* Only the addresses in deployer list can deploy a score
* Called by Governance SCORE

> Request
 
 ```json
{
  "data": {
    "method": "addDeployer",
    "params": {
      "address": "hx0123456789012345678901234567890123456789"
    }
  }
}
```

#### Parameters

| Key     | VALUE Type | Required | Description                  |
|:--------|:-----------|:---------|:-----------------------------|
| address | T_ADDRESS  | true     | address to add as a deployer |

#### Returns

`T_HASH` - txHash

### removeDeployer(address Address)

* Remove an address from deployer list
* Called by Governance SCORE
 
> Request

 ```json
{
  "data": {
    "method": "removeDeployer",
    "params": {
      "address": "hx0123456789012345678901234567890123456789"
    }
  }
}
```

#### Parameters

| Key     | VALUE Type | Required | Description                          |
|:--------|:-----------|:---------|:-------------------------------------|
| address | T_ADDRESS  | true     | address to remove from deployer list |

#### Returns

`T_HASH` - txHash

### isDeployer(address Address) bool

* Returns true if a given address is contained in deployer list

> Request

 ```json
{
  "data": {
    "method": "isDeployer",
    "params": {
      "address": "hx0123456789012345678901234567890123456789"
    }
  }
}
```

> Response
 
```json
{
  "result": "0x1"
}
``` 

#### Parameters

| Key     | VALUE Type | Required | Description      |
|:--------|:-----------|:---------|:-----------------|
| address | T_ADDRESS  | true     | address to query |

#### Returns

`T_BOOL` - boolean value representing if a given address is a deployer or not

### getDeployers() []Address

* Returns the entire addresses that are allowed to deploy a score

> Request

 ```json
{
  "data": {
    "method": "getDeployers"
  }
}
```

> Response

```json
{
  "result": [
    "hx0123456789012345678901234567890123456789",
    "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
  ]
}  
```

#### Parameters

None

#### Returns

`T_DICT` - addresses that are allowed to deploy a score

### grantValidator(address Address)

* Adds an address to validator list
* Contract address is not available
* Called by Governance SCORE

> Request
 
```json
{
  "data": {
    "method": "grantValidator",
    "params": {
      "address": "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
    }
  }
}
```

#### Parameters

| Key     | VALUE Type | Required | Description                   |
|:--------|:-----------|:---------|:------------------------------|
| address | T_ADDRESS  | true     | address to add as a validator |

#### Returns

`T_HASH` - txHash

### revokeValidator(address Address)

* Removes a validator from validator list
* Called by Governance SCORE

> Request

```json
{
  "data": {
    "method": "revokeValidator",
    "params": {
      "address": "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
    }
  }
}
```

#### Parameters

| Key     | VALUE Type | Required | Description                           |
|:--------|:-----------|:---------|:--------------------------------------|
| address | T_ADDRESS  | true     | address to remove from validator list |

#### Returns

`T_HASH` - txHash

### getValidators() []Address

* Returns the current validator list

> Request

```json
{
  "data": {
    "method": "getValidators"
  }
}
```

> Response

```json
{
  "result": [
    "hx3ece50aaa01f7c4d128c029d569dd86950c34215",
    "hxd55474243722deb1333480583cb01b38e04b90d7",
    "hx1d07c197bb3bf6131eb48969496c065e022b8f40",
    "hxbb9edb232b4519722be70a4fbf81015d0a1ed811"
  ]
}
```
 
#### Parameters

None

#### Returns

`T_LIST` - the current validator list

### setRoundLimitFactor(factor int)

* Sets a roundLimitFactor that is used for roundLimit calculation
* Called by Governance SCORE
* `RoundLimit = (len(validators) * roundLimitFactor + 2) / 3`

> Request
 
 ```json
{
  "data": {
    "method": "setRoundLimitFactor",
    "params": {
      "factor": "0x3"
    }
  }
}
```

#### Parameters

| Key    | VALUE Type | Required | Description |
|:-------|:-----------|:---------|:------------|
| factor | T_INT      | true     | -           |

#### Returns

`T_HASH` - txHash

### getRoundLimitFactor() int

* Returns the current roundLimitFactor

> Request
 
 ```json
{
  "data": {
    "method": "getRoundLimitFactor"
  }
}
```

> Response
 
```json
{
  "result": "0x3"
}
```
  
#### Parameters

None

#### Returns

`T_INT` - integer value representing roundLimitFactor

## HAVAH APIs

HAVAH-specific JSON-RPC APIs

### startRewardIssue(height int)

* Set up the block height when to issue rewards begins
* Called by Governance SCORE

> Request
 
```json
{
  "data": {
    "method": "startRewardIssue",
    "params": {
      "height": "0x64"
    }
  }
}
```

#### Parameters

| Key    | VALUE Type | Required | Description                              |
|:-------|:-----------|:---------|:-----------------------------------------|
| height | T_INT      | true     | Block height when to issue reward begins |

#### Returns

`T_HASH` - txHash

### addPlanetManager(address Address)

* Adds a specified address to PlanetManager list
* Called by Governance SCORE

> Request
 
```json
{
  "data": {
    "method": "addPlanetManager",
    "params": {
      "address": "hx0123456789012345678901234567890123456789"
    }
  }
}
```

#### Parameters

| Key     | VALUE Type | Required | Description           |
|:--------|:-----------|:---------|:----------------------|
| address | T_ADDRESS  | true     | PlanetManager address |

#### Returns

`T_HASH` - txHash

### removePlanetManager(address Address)

* Removes a specific address from PlanetManager list
* Called by Governance SCORE
 
> Request
 
```json
{
  "data": {
    "method": "removePlanetManager",
    "params": {
      "address": "hx0123456789012345678901234567890123456789"
    }
  }
}
```

#### Parameters

| Key     | VALUE Type | Required | Description           |
|:--------|:-----------|:---------|:----------------------|
| address | T_ADDRESS  | true     | PlanetManager address |

#### Returns

`T_HASH` - txHash

### isPlanetManager(address Address) bool

* Query if a specific address is a PlanetManager or not

> Request
 
```json
{
  "data": {
    "method": "isPlanetManager",
    "params": {
      "address": "hx8f21e5c54f016b6a5d5fe65486908592151a7c57"
    }
  }
}
```

> Response
 
```json
{
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
| -   | T_BOOL     | true     | false(0x0), true(0x1) |

### registerPlanet(id int, isPrivate bool, isCompany bool, owner Address, usdt int, price int)

* Registers a planet to the network
* Called by PlanetNFT SCORE

> Request
 
```json
{
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

`T_HASH` - txHash

### getPlanetInfo(id int) dict

* Returns the information on the planet specified by id

> Request
 
```json
{
  "data": {
    "method": "getPlanetInfo",
    "params": {
      "id": "0x1"
    }
  }
}
```

> Response

```json
{
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

### unregisterPlanet(id int)

* Unregisters a planet
* Unregistered planet ID can be reused
* Called by PlanetNFT SCORE

> Request
 
```json
{
  "data": {
    "method": "unregisterPlanet",
    "params": {
      "id": "0x1"
    }
  }
}
```

#### Parameters

| Key | VALUE Type | Required | Description |
|:----|:-----------|:---------|:------------|
| id  | T_INT      | true     | Planet ID   |

#### Returns

`T_HASH` - txHash

#### EventLog

* [`LostDeposited(int,int,str)`](#lostdepositedintintstr)

### setPlanetOwner(id int, owner Address)

* Changes a planet owner
* Called by PlanetNFT SCORE
 
> Request
 
```json
{
  "data": {
    "method": "setPlannetOwner",
    "params": {
      "id": "0x1",
      "owner": "hx0123456789012345678901234567890123456789"
    }
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description      |
|:------|:-----------|:---------|:-----------------|
| id    | T_INT      | true     | Planet ID        |
| owner | T_ADDRESS  | true     | New planet owner |

#### Returns

`T_HASH` - txHash

### reportPlanetWork(id int)

* PlanetManager reports a planet's work
* The network offers the rewards to a planet whose work has been reported in a term
* Only one report for a planet is available in a term
* `RewardOffered` eventlog is recorded in this transaction result
* Called by PlanetManager

> Request
 
```json
{
  "data": {
    "method": "reportPlanetWork",
    "params": {
      "id": "0x1"
    }
  }
}
```

#### Parameters

| Key | VALUE Type | Required | Description      |
|:----|:-----------|:---------|:-----------------|
| id  | T_INT      | true     | Planet ID        |

#### Returns

`T_HASH` - txHash

#### EventLog

* [`RewardOffered(int,int,int,int)`](#rewardofferedintintintint)

### claimPlanetReward(ids []int)

* Claims remaining rewards for specific planets
* Claimed rewards are transferred from `PublicTreasury` to the planet owner
* The rewards of up to `50` planets can be claimed at once.
* Called by a planet owner
 
> Request
 
```json
{
  "data": {
    "method": "claimPlanetReward",
    "params": {
      "ids": ["0x1", "0x2", "0x10"]
    }
  }
}
```

#### Parameters

| Key | VALUE Type | Required | Description                                 |
|:----|:-----------|:---------|:--------------------------------------------|
| ids | T_LIST     | true     | Planet IDs to claim rewards (max count: 50) |

#### Returns

`T_HASH` - txHash

#### EventLog

* [`RewardClaimed(Address,int,int,int)`](#rewardclaimedaddressintintint)

### getRewardInfoOf(id int) dict

* Returns the reward information on a given planet
* If a given planet id is not valid, error response is returned.

> Request

```json
{
  "data": {
    "method": "getRewardInfoOf",
    "params": {
      "id": "0x1"
    }
  }
}
```

> Response
 
 ```json
{
  "result": {
    "height": "0x3e8",
    "id": "0x1",
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
| height    | T_INT      | true     | Block height of state                                     |
| id        | T_INT      | true     | Planet ID                                                 |
| total     | T_INT      | true     | Total accumulated rewards until now                       |
| remain    | T_INT      | true     | Difference between Total Rewards and Claimed Rewards      |
| claimable | T_INT      | true     | Rewards that a planet owner can receive when claiming now |

### getRewardInfo() dict

* Returns the overall reward information
* This call returns an error response before the start of the first term
* It is not allowed to call this function via inter-call on query mode

> Request

```json
{
  "data": {
    "method": "getRewardInfo"
  }
}
```

> Response

 ```json
{
  "result": {
    "height": "0x3e8",
    "termSequence": "0x0",
    "rewardPerActivePlanet": "0xde0b6b3a7640000"
  }
}
```

#### Parameters

None

#### Returns

| Key                   | VALUE Type | Required | Description                                  |
|:----------------------|:-----------|:---------|:---------------------------------------------|
| height                | T_INT      | true     | Block height of state                        |
| termSequence          | T_INT      | false    | Sequence of a term starting with 0           |
| rewardPerActivePlanet | T_INT      | false    | Estimated reward per active planet each term |

* `rewardPerActivePlanet` does not include the fund from HooverFund SCORE
* `rewardPerActivePlanet` is zero if no active planet exists
 
### getIssueInfo() dict

* Returns the information on issue-related configuration
 
> Request
 
```json
{
  "data": {
    "method": "getIssueInfo"
  }
}
```

> Response

```json
{
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

| Key                   | VALUE Type | Required | Description                                            |
|:----------------------|:-----------|:---------|:-------------------------------------------------------|
| height                | T_INT      | true     | Block height of state                                  |
| termPeriod            | T_INT      | true     | Coin issuing term period (unit: block)                 |
| issueReductionCycle   | T_INT      | true     | issueAmount is reduced at a fixed rate every cycle     |
| issueStart            | T_INT      | false    | BlockHeight when issuing coin will begin               |
| termSequence          | T_INT      | false    | Sequence of a term starting with 0                     |

* `issueStart` is provided after the blockHeight has been set by `startRewardIssue` call
* `termSequence` is provided after the start of the first term

### getUSDTPrice() int

* Returns 1 USDT price in HVH

> Request

```json
{
  "data": {
    "method": "getUSDTPrice"
  }
}
```

> Response
 
 ```json
{
  "result": "0x8ac7230489e80000"
}
```

#### Parameters

None

#### Returns

| Key | VALUE Type | Required | Description         |
|:----|:-----------|:---------|:--------------------|
| -   | T_INT      | true     | 1 USDT price in HVH |

### setUSDTPrice(price int)

* Set 1 USDT price in HVH
* Temporary API
* Called by Governance SCORE after `revision 3`

> Request

```json
{
  "data": {
    "method": "setUSDTPrice",
    "params": {
      "price": "0x8ac7230489e80000"
    }
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description         |
|:------|:-----------|:---------|:--------------------|
| price | T_INT      | true     | 1 USDT price in HVH |

#### Returns

`T_HASH` - txHash

### fallback

* This method is called automatically when coins are transferred to `cx0000000000000000000000000000000000000000`
* To burn coins, transfer the amount of coins to burn to `cx0000000000000000000000000000000000000000`
* Every account can burn its coins

> Request

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "icx_sendTransaction",
  "params": {
    "version": "0x3",
    "from": "hxc0d4791be0b5ef67063b3c10b840fb81514db2e2",
    "to": "cx0000000000000000000000000000000000000000",
    "value": "0xde0b6b3a7640000",
    "stepLimit": "0xf4240",
    "timestamp": "0x5e3d15dcc37f7",
    "nid": "0x101",
    "signature":"ut0co6SLSzYAeellbZijrpSLZjQsi5YAyUBXPbvuwy58qeiFqnZFZISJn9NsioJUFSVf7WAx5ZUSaAkMEcY6KQE="
  }
}
```

#### Parameters

N/A

#### Returns

`T_HASH` - txHash

#### EventLog

* [`Burned(Address,int,int)`](#burnedaddressintint)

### setPrivateClaimableRate(numerator int, denominator int)

* Sets claimable rate for a private planet
* Called by Governance SCORE

> Request

```json
{
  "data": {
    "method": "setPrivateClaimableRate",
    "params": {
      "numerator": "0x1",
      "denominator": "0x18"
    }
  }
}
```

#### Parameters

| Key         | VALUE Type | Required | Description                         |
|:------------|:-----------|:---------|:------------------------------------|
| numerator   | T_INT      | true     | Numerator of PrivateClaimableRate   |
| denominator | T_INT      | true     | Denominator of PrivateClaimableRate |

* 0 <= numerator <= 10000
* 0 < denominator <= 10000
* numerator <= denominator

#### Returns

`T_HASH` - txHash

### getPrivateClaimableRate() dict

* Returns claimable rate for a private planet
* If PrivateClaimableRate has not changed, this call returns numerator=0, denominator=24 as default

> Request

```json
{
  "data": {
    "method": "getPrivateClaimableRate"
  }
}
```

> Response
 
```json
{
  "result": {
    "numerator": "0x0",
    "denominator": "0x18"
  }
}
```

#### Parameters

None

#### Returns

`T_DICT` - numerator and denominator representing PrivateClaimableRate

### getRewardInfosOf(ids []int) dict

* Returns the reward information on given planets
* If it is failed to get reward information on some planets, the only corresponding information will be replaced with `null`.
* Maximum number of ids: `50`
* Since `revision 2`

> Request

```json
{
  "data": {
    "method": "getRewardInfosOf",
    "params": {
      "ids": ["0x1", "0x2", "0x3"]
    }
  }
}
```

> Response
 
```json
{
  "result": {
    "height": "0x20e",
    "rewardInfos": [
      {
        "id": "0x1",
        "total": "0xb61cb183c4b7e1800000",
        "remain": "0xb61cb183c4b7e1800000",
        "claimable": "0xb61cb183c4b7e1800000"
      },
      null,
      {
        "id": "0x3",
        "total": "0xb61cb183c4b7e1800000",
        "remain": "0x0",
        "claimable": "0x0"
      }
    ]
  }
}
```

#### Parameters

| Key | VALUE Type | Required | Description |
|:----|:-----------|:---------|:------------|
| id  | T_INT      | true     | Planet ID   |

#### Returns

| Key         | VALUE Type      | Required | Description                              |
|:------------|:----------------|:---------|:-----------------------------------------|
| height      | T_INT           | true     | Block height of state                    |
| rewardInfos | []T_REWARD_INFO | true     | Reward information list of given planets |

> T_REWARD_INFO
 
 | Key       | VALUE Type | Required | Description                                               |
|:----------|:-----------|:---------|:----------------------------------------------------------|
| id        | T_INT      | true     | Planet ID                                                 |
| total     | T_INT      | true     | Total accumulated rewards until now                       |
| remain    | T_INT      | true     | Difference between Total Rewards and Claimed Rewards      |
| claimable | T_INT      | true     | Rewards that a planet owner can receive when claiming now |

### withdrawLostTo(to Address)

* Withdraw all lost coins to a given `to` address
* Only EOA is allowed for `to` address
* Called by Governance SCORE
* Since `revision 2`

> Request

```json
{
  "data": {
    "method": "withdrawLostTo",
    "params": {
      "to": "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
    }
  }
}
```

#### Parameters

| Key | VALUE Type | Required | Description                                |
|:----|:-----------|:---------|:-------------------------------------------|
| to  | T_ADDRESS  | true     | Address where lost coins will be withdrawn |

#### Returns

`T_HASH` - txHash

#### EventLog

* [`LostWithdrawn(Address,int)`](#lostwithdrawnaddressint)
 
### getLost() int

* Returns the amount of lost coins
* Since `revision 2`

> Request

```json
{
  "data": {
    "method": "getLost"
  }
}
```

> Response
 
 ```json
{
  "result": "0x8ac7230489e80000"
}
```

#### Parameters

None

#### Returns

| Key | VALUE Type | Required | Description          |
|:----|:-----------|:---------|:---------------------|
| -   | T_INT      | true     | Amount of lost coins |

### registerValidator(owner Address, nodePublicKey bytes, grade string, name string)

* Registers a validator
* Called by Governance SCORE
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "registerValidator",
    "params": {
      "owner": "hx3ece50aaa01f7c4d128c029d569dd86950c34215",
      "nodePublicKey": "0x03921d018c1302f9be67e0aca8e583f8970af127e18727d68af4e7caa81c8eb2b1",
      "grade": "sub",
      "name": "ABC node"  
    }
  }
}
```

#### Parameters

| Key           | VALUE Type | Required | Description                         |
|:--------------|:-----------|:---------|:------------------------------------|
| owner         | T_ADDRESS  | true     | Node owner address                  |
| nodePublicKey | T_BYTES    | true     | Node publicKey in compressed format |
| grade         | T_STRING   | true     | `sub`, `main`                       | 
| name          | T_STRING   | true     | Node name                           |

* Only publicKeys in compressed format are accepted

#### Returns

`T_HASH` - txHash

### unregisterValidator(owner Address)

* Unregisters a validator
* Called by Governance SCORE or validator owner
* Once a validator has been unregistered, the node cannot be a validator anymore
* Since `revision 4`

> Request
 
```json
{
  "data": {
    "method": "unregisterValidator",
    "params": {
      "owner": "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
    } 
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description        |
|:------|:-----------|:---------|:-------------------|
| owner | T_ADDRESS  | true     | Node owner address |

#### Returns

`T_HASH` - txHash

#### EventLog

* [`ActiveValidatorRemoved(Address,Address,string)`](#activevalidatorremovedaddressaddressstring)
* [`ActiveValidatorAdded(Address,Address)`](#activevalidatoraddedaddressaddress)
 
### setBlockVoteCheckParameters(period int, allowance int)

* Sets `blockVoteCheckPeriod` and `nonVoteAllowance` parameters used after decentralization
* Changed parameters are applied at the beginning of next term.
* Called by Governance SCORE
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "setBlockVoteCheckParameters",
    "params": {
      "period": "0x64",
      "allowance": "0x5"
    }
  }
}
```

#### Parameters

| Key       | VALUE Type | Required | Description                                                              |
|:----------|:-----------|:---------|:-------------------------------------------------------------------------|
| period    | T_INT      | true     | block vote checking period in blocks                                     |
| allowance | T_INT      | true     | A validator gets penalized if its nonVotes is larger than this allowance |

#### Returns

`T_HASH` - txHash

### getBlockVoteCheckParameters() dict

* Returns `blockVoteCheckPeriod` and `nonVoteAllowance` parameters
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "getBlockVoteCheckParameters"
  }
}
```

> Response

 ```json
{
  "result": {
    "blockVoteCheckPeriod": "0x1e",
    "nonVoteAllowance": "0x14"
  }
}
```

#### Parameters

None

#### Returns

| Key                  | VALUE Type | Required | Description                                                              |
|:---------------------|:-----------|:---------|:-------------------------------------------------------------------------|
| blockVoteCheckPeriod | T_INT      | true     | block vote checking period in blocks                                     |
| nonVoteAllowance     | T_INT      | true     | A validator gets penalized if its nonVotes is larger than this allowance |

### getNetworkStatus() dict

* Returns network status after decentralization
* If it is not decentralized, an error response is returned
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "getNetworkStatus"
  }
}
```

> Response

 ```json
{
  "result": {
    "height": "0x3e8",
    "termStart": "0x3e0",
    "mode": "0x1",
    "blockVoteCheckPeriod": "0x1e",
    "nonVoteAllowance": "0x14",
    "activeValidatorCount": "0xa"
  }
}
```

#### Parameters

None

#### Returns

| Key                  | VALUE Type | Required | Description                                                            |
|:---------------------|:-----------|:---------|:-----------------------------------------------------------------------|
| height               | T_INT      | true     | Block height of state                                                  |
| termStart            | T_INT      | true     | Block height when this term starts                                     |
| mode                 | T_INT      | true     | Network mode. init(0), decentralization(1)                             |
| blockVoteCheckPeriod | T_INT      | true     | Period in blocks of block vote check                                   |
| nonVoteAllowance     | T_INT      | true     | Validator gets penalized if its nonVotes is larger than this allowance |
| activeValidatorCount | T_INT      | true     | Maximum number of active validators                                    |

### setValidatorInfo(values []{key string, value string})

* Changes some properties of a validator info
* Owners can update as many fields in validator information as they want at once
* Called by validator owner
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "setValidatorInfo",
    "params": [
      {
        "key": "name",
        "value": "ABC Node"
      },
      {
        "key": "url",
        "value": "https://www.worldbestnode.com/details.json"
      }
    ]      
  }
}
```

#### Parameters

| Key    | VALUE Type    | Required | Description                                 |
|:-------|:--------------|:---------|:--------------------------------------------|
| values | []T_KEY_VALUE | true     | Key/value pairs to change in validator info |

> T_KEY_VALUE

| Key   | VALUE Type | Required | Description               |
|:------|:-----------|:---------|:--------------------------|
| key   | T_STRING   | true     | `name`, `url`             |
| value | T_STRING   | true     | New value for a given key |

* Maximum name length: `100` in bytes
* Maximum url length: `200` in bytes
* Additional values for key field would be added in the future

#### Returns

`T_HASH` - txHash

### getValidatorInfo(owner Address) dict

* Returns the information on the validator indicated by owner
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "getValidatorInfo",
    "params": {
      "owner": "hx0123456789012345678901234567890123456789"
    }
  }
}
```

> Response

 ```json
{
  "result": {
    "height": "0x3e8",
    "owner": "hx0123456789012345678901234567890123456789",
    "nodePublicKey": "0x03921d018c1302f9be67e0aca8e583f8970af127e18727d68af4e7caa81c8eb2b1",
    "node": "hx2563e6542d310a4e09f5f2851f2c497bdce45ac0",
    "grade": "sub",
    "name": "Excellent node",
    "url": "https://www.excellentnode.com/details.json"
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description     |
|:------|:-----------|:---------|:----------------|
| owner | T_ADDRESS  | true     | Validator owner |

#### Returns

| Key           | VALUE Type | Required | Description                                    |
|:--------------|:-----------|:---------|:-----------------------------------------------|
| height        | T_INT      | true     | Block height of state                          |
| owner         | T_ADDRESS  | true     | Validator owner (immutable)                    |
| nodePublicKey | T_BYTES    | true     | Compressed node publicKey                      |
| node          | T_ADDRESS  | true     | Node address derived from nodePublicKey        |
| grade         | T_STRING   | true     | `sub`, `main`                                  |
| name          | T_STRING   | true     | Validator name                                 |
| url           | T_STRING   | true     | URL to return validator details in json format |

* The json format of validator details will be described in a separate document.

### setNodePublicKey(pubKey bytes)

* Change an existing nodePublicKey to new one
* Used nodePublicKeys are not allowed, including those of other owners
* Called by validator owner
* Since `revision 4`
 
> Request

```json
{
  "data": {
    "method": "setNodePublicKey",
    "params": {
      "pubKey": "0x03921d018c1302f9be67e0aca8e583f8970af127e18727d68af4e7caa81c8eb2b1"
    }
  }
}
```

#### Parameters

| Key    | VALUE Type | Required | Description                        |
|:-------|:-----------|:---------|:-----------------------------------|
| pubKey | T_BYTES    | true     | nodePublicKey in compressed format |

* Only publicKeys in compressed format are accepted

#### Returns

`T_HASH` - txHash

#### EventLog

* [`ActiveValidatorRemoved(Address,Address,string)`](#activevalidatorremovedaddressaddressstring)
* [`ActiveValidatorAdded(Address,Address)`](#activevalidatoraddedaddressaddress)
 
### enableValidator(owner Address)

* Enable a disabled validator indicated by owner, resetting its nonVotes to 0
* If it is called by Governance SCORE, then the enableCount of validator will be reverted to initial value
* Initial enableCount: `3`
* Called by Governance SCORE or validator owner
* Since `revision 4` 
 
> Request

```json
{
  "data": {
    "method": "enableValidator",
    "params": {
      "owner": "hx0123456789012345678901234567890123456789"
    }
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description     |
|:------|:-----------|:---------|:----------------|
| owner | T_ADDRESS  | true     | Validator owner |

#### Returns

`T_HASH` - txHash

### getValidatorStatus(owner Address) dict

* Returns the validator status indicated by owner
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "getValidatorStatus",
    "params": {
      "owner": "hx0123456789012345678901234567890123456789"
    }
  }
}
```

> Response

```json
{
  "result": {
    "height": "0x3e8",
    "flags": "0x0",
    "nonVotes": "0x0",
    "enableCount": "0x3"
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description     |
|:------|:-----------|:---------|:----------------|
| owner | T_ADDRESS  | true     | Validator owner |

#### Returns

| Key         | VALUE Type | Required | Description                                                          |
|:------------|:-----------|:---------|:---------------------------------------------------------------------|
| height      | T_INT      | true     | Block height of state                                                |
| flags       | T_INT      | true     | Bitwise flags: disabled(1), disqualified(2)                          |
| nonVotes    | T_INT      | true     | Number of times that a validator did not participate in a block vote |
| enableCount | T_INT      | true     | Number of times that a validator can be enabled                      |

#### Returns

`T_HASH` - txHash

### setActiveValidatorCount(count int)

* Sets the maximum number of active validators participating in block validation
* Maximum number of active validators: `1000`
* Called by Governance SCORE
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "setActiveValidatorCount",
    "params": {
      "count": "0xa"
    }
  }
}
```

#### Parameters

| Key   | VALUE Type | Required | Description                                                           |
|:------|:-----------|:---------|:----------------------------------------------------------------------|
| count | T_INT      | true     | Maximum number of active validators participating in block validation |

#### Returns

`T_HASH` - txHash

### getActiveValidatorCount() int

* Returns the maximum number of active validators to take effect in the next term
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "getActiveValidatorCount"
  }
}
```

> Response

```json
{
  "result": "0xa"
}
```

#### Parameters

None

#### Returns

| Key | VALUE Type | Required | Description                                                         |
|:----|:-----------|:---------|:--------------------------------------------------------------------|
| -   | T_INT      | true     | Maximum number of active validators to take effect in the next term |

### getValidatorsOf(grade string) dict

* Returns the owner addresses of registered validators filtered by `grade`
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "getValidatorsOf",
    "params": {
      "grade": "sub"
    }
  }
}
```

> Response

```json
{
  "result": {
    "height": "0x3e8",
    "grade": "sub",
    "validators": [
      "hxcb90235946c8ab02e24a03b898e0994ddb78d13a",
      "hxe3979d6118e437cff510ac6c989fe7c48162a32f",
      "hxe6710e03332914d97bb7dff3ee6831a03cd0a7ae",
      "hxcb35f82a3a943e040ae2b9ab2baa2118781b2bc9",
      "hxf5be47d975a2a71a22156713ab58ec5e2338ba34"
    ]
  }
}
```

#### Parameters

None

#### Returns

| Key        | VALUE Type  | Required | Description                                                 |
|:-----------|:------------|:---------|:------------------------------------------------------------|
| height     | T_INT       | true     | Block height of state                                       |
| grade      | T_STRING    | true     | Grade of validators: `sub`, `main`, `all`                   |
| validators | []T_ADDRESS | true     | Owner addresses of registered validators with a given grade |

* `all` means all validators regardless of their grades
* Unregistered validators are not contained in `validators` value

### getValidatorsInfo(dataType string) dict

* Returns the information on registered validators
* The information contains the results of [getValidatorInfo](#getvalidatorinfoowner-address-dict) or [getValidatorStatus](#getvalidatorstatusowner-address-dict)
* Since `revision 4`

> Request

```json
{
  "data": {
    "method": "getValidatorsInfo",
    "params": {
      "dataType": "all"
    }
  }
}
```

> Response

```json
{
  "result": {
    "height": "0x3e8",
    "grade": "all",
    "validators": [
      {
        "enableCount": "0x3",
        "flags": "0x0",
        "grade": "main",
        "name": "main-0",
        "node": "hxa8df82e93e8a9cd5325e37289bcd0fbc0a8b4e5e",
        "nodePublicKey": "0x0335a46022ed9c8e3ba0e82951d15062f63d15d778f6b3a400bfd183ce322b80a1",
        "nonVotes": "0x0",
        "owner": "hxb6dc91d57ef0c94a29a5e5f709db323853022093",
        "url": "https://www.havah.com/details.json"
      },
      {
        "enableCount": "0x2",
        "flags": "0x0",
        "grade": "sub",
        "name": "sub-0",
        "node": "hxb9e093f93e8a9cd5325e37289bcd0fbc0b9c5f6f",
        "nodePublicKey": "0x0293b57132fc9c8e3ba0e82951d15062f63d15d778f6b3a400bfd183ce402c91b2",
        "nonVotes": "0x1",
        "owner": "hxc7eda2e68fc0c94a29a5e5f709db3239641331a4",
        "url": "https://www.example.com/details.json"
      }
    ]
  }
}
```

#### Parameters

| Key      | VALUE Type | Required | Description             |
|:---------|:-----------|:---------|:------------------------|
| dataType | T_STRING   | true     | `info`, `status`, `all` |

#### Returns

| Key        | VALUE Type         | Required | Description                                  |
|:-----------|:-------------------|:---------|:---------------------------------------------|
| height     | T_INT              | true     | Block height of state                        |
| validators | []T_VALIDATOR_INFO | true     | List of information of registered validators |

> T_VALIDATOR_INFO
 
* Contains results of [getValidatorInfo](#getvalidatorinfoowner-address-dict) or [getValidatorStatus](#getvalidatorstatusowner-address-dict) depending on `dataType`
 
## EventLogs

HAVAH records the following eventLogs:

### Transfer(Address,Address,int)

* Logged only when a SCORE transfers coins to an address
* This eventLog is not recorded when an EOA transfers coins
* ScoreAddress: `cx0000000000000000000000000000000000000000`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "Transfer(Address,Address,int)",
    "cx0f4dbedd2b5cf3323ea23371b84576bcc438140f",
    "hx0123456789012345678901234567890123456789"
  ],
  "data":[
    "0xde0b6b3a7640000"
  ]
}
```

| Key    | VALUE Type | Indexed | Description                 |
|:-------|:-----------|:--------|:----------------------------|
| from   | T_ADDRESS  | true    | from address                |
| to     | T_ADDRESS  | true    | to address                  |
| amount | T_INT      | false   | Amount of coins transferred |
 
### Issued(int,int,int)

* Logged on BaseTx when issuing coins each term start
* ScoreAddress: `cx0000000000000000000000000000000000000000`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "Issued(int,int,int)"
  ],
  "data":[
    "0x1",
    "0xde0b6b3a7640000",
    "0x308501e99f05f71326a0914"
  ]
}
```

| Key          | VALUE Type | Indexed | Description                     |
|:-------------|:-----------|:--------|:--------------------------------|
| termSequence | T_INT      | false   | Term sequence starting with 0   |
| amount       | T_INT      | false   | Amount of coins issued          |
| totalSupply  | T_INT      | false   | totalSupply after issuing coins |

### Burned(Address,int,int)

* Logged when [burning coins](#fallback)
* ScoreAddress: `cx0000000000000000000000000000000000000000`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "Burned(Address,int,int)",
    "cx0000000000000000000000000000000000000000"
  ],
  "data":[
    "0xde0b6b3a7640000",
    "0x308501e99f05f71326a0914"
  ]
}
```

| Key         | VALUE Type | Indexed | Description                     |
|:------------|:-----------|:--------|:--------------------------------|
| owner       | T_ADDRESS  | true    | Address of burned coin owner    |
| amount      | T_INT      | false   | Amount of burned coins          |
| totalSupply | T_INT      | false   | totalSupply after burning coins |

### HooverRefilled(int,int,int)

* Logged on BaseTx when hooverFund is refilled each term
* ScoreAddress: `cx0000000000000000000000000000000000000000`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "HooverRefilled(int,int,int)"
  ],
  "data":[
    "0xde0b6b3a7640000",
    "0x38e8f7792d79767800000",
    "0x422ca8b0a00a425000000"
  ]
}
```

| Key                    | VALUE Type | Indexed | Description                            |
|:-----------------------|:-----------|:--------|:---------------------------------------|
| amount                 | T_INT      | false   | Amount of refilled funds               |
| hooverBalance          | T_INT      | false   | HooverFund balance after refilled      |
| sustainableFundBalance | T_INT      | false   | SustainableFund balance after refilled |

### RewardOffered(int,int,int,int)

* Logged when [`reportPlanetWork`](#reportplanetworkid-int) is called
* ScoreAddress: `cx0000000000000000000000000000000000000000`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "RewardOffered(int,int,int,int)"
  ],
  "data":[
    "0x12",
    "0x64",
    "0xde0b6b3a7640000",
    "0x16345785d8a0000"
  ]
}
```

| Key              | VALUE Type | Indexed | Description                                                              |
|:-----------------|:-----------|:--------|:-------------------------------------------------------------------------|
| termSequence     | T_INT      | false   | Term sequence starting with 0                                            |
| id               | T_INT      | false   | Planet ID                                                                |
| rewardWithHoover | T_INT      | false   | Rewards in HVH that the planet gets including an subsidy from HooverFund |
| hooverRequest    | T_INT      | false   | Subsidy from HooverFund                                                  |

### RewardClaimed(Address,int,int,int)

* Logged when [`claimPlanetReward`](#claimplanetrewardids-int) is called
* ScoreAddress: `cx0000000000000000000000000000000000000000`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "RewardClaimed(Address,int,int,int)",
    "hx0123456789012345678901234567890123456789"
  ],
  "data":[
    "0x12",
    "0x64",
    "0xde0b6b3a7640000"
  ]
}
```

| Key          | VALUE Type | Indexed | Description                   |
|:-------------|:-----------|:--------|:------------------------------|
| owner        | T_ADDRESS  | true    | Planet owner claiming rewards |
| termSequence | T_INT      | false   | Term sequence starting with 0 |
| id           | T_INT      | false   | Planet ID                     |
| amount       | T_INT      | false   | Claimed reward amount         |

### LostDeposited(int,int,str)

* Logged when [`unregisterPlanet`](#unregisterplanetid-int) is called
* If all rewards for an unregistered planet have been already claimed, this event is not logged as there is no lost coin
* ScoreAddress: `cx0000000000000000000000000000000000000000`
* Since `revision 2`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "LostDeposited(int,int,str)"
  ],
  "data":[
    "0xde0b6b3a7640000",
    "0x16345785d8a00000",
    "PlanetUnregistered(id=1)"
  ]
}
```

| Key       | VALUE Type | Indexed | Description                                                 |
|:----------|:-----------|:--------|:------------------------------------------------------------|
| lostDelta | T_INT      | false   | Amount of lost coins newly added                            |
| lostTotal | T_INT      | false   | Amount of accumulated lost coins remaining deposited so far |
| reason    | T_STRING   | false   | Reason for new lost coins                                   |

### LostWithdrawn(Address,int)

* Logged when [`withdrawLostTo`](#withdrawlosttoto-address) is called
* If the amount of lost coins to withdraw is 0, no event will be logged
* ScoreAddress: `cx0000000000000000000000000000000000000000`
* Since `revision 2`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "LostWithdrawn(Address,int)"
  ],
  "data":[
    "hx3ece50aaa01f7c4d128c029d569dd86950c34215",
    "0x16345785d8a00000"
  ]
}
```

| Key    | VALUE Type | Indexed | Description                    |
|:-------|:-----------|:--------|:-------------------------------|
| to     | T_ADDRESS  | false   | Recipient to get lost coins    |
| amount | T_INT      | false   | Amount of lost coins withdrawn |

### TermStarted(int,int,int)

* Logged at the beginning of each term
* This event is located in base transaction result
* ScoreAddress: `cx0000000000000000000000000000000000000000`
* Since `revision 2`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "TermStarted(int,int,int)"
  ],
  "data":[
    "0x1",
    "0xa",
    "0xde0b6b3a7640000"
  ]
}
```

| Key                   | VALUE Type | Indexed | Description                                            |
|:----------------------|:-----------|:--------|:-------------------------------------------------------|
| termSequence          | T_INT      | false   | Sequence of a term starting with 0                     |
| planetCount           | T_INT      | false   | Number of active planets at the beginning of each term |
| rewardPerActivePlanet | T_INT      | false   | Estimated reward per active planet each term           |

### Decentralized(int)

* Logged only once at the beginning of decentralization
* Located in base transaction result
* ScoreAddress: `cx0000000000000000000000000000000000000000`
* Since `revision 4` 

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "Decentralized(int)"
  ],
  "data":[
    "0xa"
  ]
}
```

| Key                  | VALUE Type | Indexed | Description                         |
|:---------------------|:-----------|:--------|:------------------------------------|
| activeValidatorCount | T_INT      | false   | Maximum number of active validators |

### ActiveValidatorAdded(Address,Address)

* Logged when a validator was added to active validator set
* ScoreAddress: `cx0000000000000000000000000000000000000000`
* Since `revision 4`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "ActiveValidatorAdded(Address,Address)",
    "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
  ],
  "data":[
    "hx0123456789012345678901234567890123456789"
  ]
}
```

| Key   | VALUE Type | Indexed | Description             |
|:------|:-----------|:--------|:------------------------|
| owner | T_ADDRESS  | true    | Validator owner address |
| node  | T_ADDRESS  | false   | Validator node address  |

### ActiveValidatorRemoved(Address,Address,string)

* Logged when an active validator was removed from active validator set
* ScoreAddress: `cx0000000000000000000000000000000000000000`
* Since `revision 4`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "ActiveValidatorRemoved(Address,Address)",
    "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
  ],
  "data":[
    "hx0123456789012345678901234567890123456789",
    "penalized"
  ]
}
```

| Key    | VALUE Type | Indexed | Description                                               |
|:-------|:-----------|:--------|:----------------------------------------------------------|
| owner  | T_ADDRESS  | true    | Validator owner address                                   |
| node   | T_ADDRESS  | false   | Validator node address                                    |
| reason | T_STRING   | false   | `penalized`, `termchange`, `unregistered`, `pubkeychange` |

### ActiveValidatorPenalized(Address,Address)

* Logged when a validator got penalized
* Located in base transaction result
* ScoreAddress: `cx0000000000000000000000000000000000000000`
* Since `revision 4`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "ActiveValidatorPenalized(Address,Address)",
    "hx3ece50aaa01f7c4d128c029d569dd86950c34215"
  ],
  "data":[
    "hx0123456789012345678901234567890123456789"
  ]
}
```

| Key    | VALUE Type | Indexed | Description               |
|:-------|:-----------|:--------|:--------------------------|
| owner  | T_ADDRESS  | true    | Validator owner address   |
| node   | T_ADDRESS  | false   | Validator node address    |

### ActiveValidatorCountChanged(int,int)

* Logged when the number of active validator set has been changed
* ScoreAddress: `cx0000000000000000000000000000000000000000`
* Since `revision 4`

```json
{
  "scoreAddress": "cx0000000000000000000000000000000000000000",
  "indexed":[
    "ActiveValidatorCountChanged(int,int)"
  ],
  "data":[
    "0x9",
    "0xa"
  ]
}
```

| Key       | VALUE Type | Indexed | Description                        |
|:----------|:-----------|:--------|:-----------------------------------|
| old count | T_INT      | false   | Number of old active validator set |
| new count | T_INT      | false   | Number of new active validator set |
