package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.jsonrpc.RpcArray;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.ResultTimeoutException;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;

import java.io.IOException;
import java.math.BigInteger;

public class ChainScore extends foundation.icon.test.score.ChainScore {
    public ChainScore(TransactionHandler txHandler) {
        super(txHandler);
    }

    public TransactionResult setScoreOwner(Wallet wallet, Address score, Address owner) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("score", new RpcValue(score))
                .put("owner", new RpcValue(owner))
                .build();
        return invokeAndWaitResult(wallet, "setScoreOwner", params, null, foundation.icon.test.common.Constants.DEFAULT_STEPS);
    }

 /// ----------- apis for HAVAH ----------------
    public BigInteger getUSDTPrice() throws IOException {
        var v = call("getUSDTPrice", null);
        return call("getUSDTPrice", null).asInteger();
    }

    public Bytes invokeReportPlanetWork(Wallet wallet, BigInteger id) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("id", new RpcValue(id))
                .build();
        return invoke(wallet, "reportPlanetWork", params, null, Constants.DEFAULT_STEPS);
    }

    public TransactionResult reportPlanetWork(Wallet wallet, BigInteger id) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("id", new RpcValue(id))
                .build();
        return invokeAndWaitResult(wallet, "reportPlanetWork", params, null, Constants.DEFAULT_STEPS);
    }

    public boolean isPlanetManager(Address address) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();
        return call("isPlanetManager", params).asBoolean();
    }

    public Bytes invokeClaimPlanetReward(Wallet wallet, BigInteger[] ids) throws IOException, ResultTimeoutException {
        var array = new RpcArray.Builder();
        for (var p : ids) {
            array.add(new RpcValue(p));
        }
        RpcObject params = new RpcObject.Builder()
                .put("ids", array.build())
                .build();
        return invoke(wallet, "claimPlanetReward", params);
    }

    public TransactionResult claimPlanetReward(Wallet wallet, BigInteger[] ids) throws IOException, ResultTimeoutException {
        var array = new RpcArray.Builder();
        for (var p : ids) {
            array.add(new RpcValue(p));
        }
        RpcObject params = new RpcObject.Builder()
                .put("ids", array.build())
                .build();
        return invokeAndWaitResult(wallet, "claimPlanetReward", params);
    }

    public RpcObject getPlanetInfo(BigInteger planetId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("id", new RpcValue(planetId))
                .build();
        return call("getPlanetInfo", params).asObject();
    }

    public RpcObject getIssueInfo() throws IOException {
        RpcObject params = new RpcObject.Builder()
                .build();
        return call("getIssueInfo", params).asObject();
    }

    public RpcObject getRewardInfoOf(BigInteger planetId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("id", new RpcValue(planetId))
                .build();
        return call("getRewardInfoOf", params).asObject();
    }

    ////
    public TransactionResult setMaxStepLimit(Wallet wallet, String type, BigInteger cost) throws ResultTimeoutException, IOException {
        RpcObject params = new RpcObject.Builder()
                .put("contextType", new RpcValue(type))
                .put("limit", new RpcValue(cost))
                .build();
        return invokeAndWaitResult(wallet, "setMaxStepLimit", params);
    }
}

