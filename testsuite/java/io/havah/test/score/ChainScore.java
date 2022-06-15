package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.jsonrpc.RpcArray;
import foundation.icon.icx.transport.jsonrpc.RpcItem;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.ResultTimeoutException;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;
import score.Context;

import java.io.IOException;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

public class ChainScore extends Score {
    public boolean isDeployerWhiteListEnabled() throws IOException {
        return isDeployerWhiteListEnabled(this.getServiceConfig());
    }

    public boolean isDeployer(Address address) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();
        return call("isDeployer", params).asBoolean();
    }

    public List<Address> getDeployers() throws IOException {
        List<Address> list = new ArrayList<>();
        RpcArray items = call("getDeployers", null).asArray();
        for (RpcItem item : items) {
            list.add(item.asAddress());
        }
        return list;
    }

    private static final int CONFIG_AUDIT = 0x2;
    private static final int CONFIG_DEPLOYER_WHITE_LIST = 0x4;

    public ChainScore(TransactionHandler txHandler) {
        super(txHandler, Constants.CHAINSCORE_ADDRESS);
    }

    public int getRevision() throws IOException {
        return call("getRevision", null).asInteger().intValue();
    }

    public BigInteger getStepPrice() throws IOException {
        return call("getStepPrice", null).asInteger();
    }

    public int getServiceConfig() throws IOException {
        return call("getServiceConfig", null).asInteger().intValue();
    }

    public static boolean isAuditEnabled(int config) {
        return (config & CONFIG_AUDIT) != 0;
    }

    public boolean isAuditEnabled() throws IOException {
        return isAuditEnabled(this.getServiceConfig());
    }

    public static boolean isDeployerWhiteListEnabled(int config) {
        return (config & CONFIG_DEPLOYER_WHITE_LIST) != 0;
    }

    public RpcObject getScoreStatus(Address address) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();
        return call("getScoreStatus", params).asObject();
    }

    public TransactionResult disableScore(Wallet wallet, Address address)
            throws IOException, ResultTimeoutException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();
        return invokeAndWaitResult(wallet, "disableScore", params, null, Constants.DEFAULT_STEPS);
    }

    public TransactionResult enableScore(Wallet wallet, Address address)
            throws IOException, ResultTimeoutException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();
        return invokeAndWaitResult(wallet, "enableScore", params, null, Constants.DEFAULT_STEPS);
    }

 /// ----------- apis for HAVAH ----------------
    public BigInteger getUSDTPrice() throws IOException {
        var v = call("getUSDTPrice", null);
        System.out.println("getUSDTPrice (" + v + ")");
        return call("getUSDTPrice", null).asInteger();
    }

    public TransactionResult reportPlanetWork(Wallet wallet, BigInteger id) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(id))
                .build();
        return invokeAndWaitResult(wallet, "reportPlanetWork", params, null, Constants.DEFAULT_STEPS);
    }

    public boolean isPlanetManager(Address address) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();
        return call("isPlanetManager", params).asBoolean();
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

    public RpcObject getRewardInfo(BigInteger planetId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("id", new RpcValue(planetId))
                .build();
        return call("getRewardInfo", params).asObject();
    }
}
