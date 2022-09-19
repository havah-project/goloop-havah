package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.ResultTimeoutException;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;

import java.io.IOException;
import java.math.BigInteger;
import java.util.Map;

public class GovScore extends Score {
//    private final Wallet governorWallet;
    private final ChainScore chainScore;

    public static class Fee {
        Map<String, BigInteger> stepCosts;
        Map<String, BigInteger> stepMaxLimits;
        BigInteger stepPrice;
    }

    public static final String[] stepCostTypes = {
            "default",
            "contractCall",
            "contractCreate",
            "contractUpdate",
            "contractDestruct",
            "contractSet",
            "get",
            "set",
            "replace",
            "delete",
            "input",
            "eventLog",
            "apiCall"
    };

    public GovScore(TransactionHandler txHandler) {
        super(txHandler, Constants.GOV_ADDRESS);
//        this.governorWallet = txHandler.getChain().governorWallet;
        this.chainScore = new ChainScore(txHandler);
    }

//    private Wallet getWallet() {
//        return this.governorWallet;
//    }

    @Override
    public TransactionResult invokeAndWaitResult(Wallet wallet, String method, RpcObject params)
            throws ResultTimeoutException, IOException {
        return super.invokeAndWaitResult(wallet, method, params, BigInteger.ZERO, Constants.DEFAULT_STEPS);
    }

//    public TransactionResult setRevision(int code) throws Exception {
//        RpcObject params = new RpcObject.Builder()
//                .put("code", new RpcValue(BigInteger.valueOf(code)))
//                .build();
//        return invokeAndWaitResult(getWallet(), "setRevision", params);
//    }
//
//    public void setStepPrice(BigInteger price) throws Exception {
//        RpcObject params = new RpcObject.Builder()
//                .put("price", new RpcValue(price))
//                .build();
//        invokeAndWaitResult(getWallet(), "setStepPrice", params);
//    }
//
//    public void setStepCost(String type, BigInteger cost) throws ResultTimeoutException, IOException {
//        RpcObject params = new RpcObject.Builder()
//                .put("type", new RpcValue(type))
//                .put("cost", new RpcValue(cost))
//                .build();
//        invokeAndWaitResult(getWallet(), "setStepCost", params);
//    }
//
//    public TransactionResult setMaxStepLimit(String type, BigInteger cost) throws ResultTimeoutException, IOException {
//        RpcObject params = new RpcObject.Builder()
//                .put("contextType", new RpcValue(type))
//                .put("limit", new RpcValue(cost))
//                .build();
//        return invokeAndWaitResult(getWallet(), "setMaxStepLimit", params);
//    }
//
//    public boolean isAuditEnabledOnly() throws IOException {
//        int config = chainScore.getServiceConfig();
//        return ChainScore.isAuditEnabled(config) && !ChainScore.isDeployerWhiteListEnabled(config);
//    }
//
//    public TransactionResult acceptScore(Bytes txHash) throws ResultTimeoutException, IOException {
//        RpcObject params = new RpcObject.Builder()
//                .put("txHash", new RpcValue(txHash))
//                .build();
//        return invokeAndWaitResult(getWallet(), "acceptScore", params);
//    }
//
//    public TransactionResult rejectScore(Bytes txHash) throws ResultTimeoutException, IOException {
//        RpcObject params = new RpcObject.Builder()
//                .put("txHash", new RpcValue(txHash))
//                .build();
//        return invokeAndWaitResult(getWallet(), "rejectScore", params);
//    }
//
//    public Map<String, BigInteger> getStepCosts() throws Exception {
//        RpcItem rpcItem = this.chainScore.call("getStepCosts", null);
//        Map<String, BigInteger> map = new HashMap<>();
//        for(String type : stepCostTypes) {
//            map.put(type, rpcItem.asObject().getItem(type).asInteger());
//        }
//        return map;
//    }
//
//    public void setStepCosts(Map<String, BigInteger> map)
//            throws ResultTimeoutException, TransactionFailureException, IOException {
//        List<Bytes> list = new LinkedList<>();
//        for(String type : map.keySet()) {
//            RpcObject params = new RpcObject.Builder()
//                    .put("type", new RpcValue(type))
//                    .put("cost", new RpcValue(map.get(type)))
//                    .build();
//            Bytes txHash = invoke(getWallet(), "setStepCost", params);
//            list.add(txHash);
//        }
//        for(Bytes txHash : list) {
//            TransactionResult result = getResult(txHash);
//            if (!Constants.STATUS_SUCCESS.equals(result.getStatus())) {
//                throw new TransactionFailureException(result.getFailure());
//            }
//        }
//    }
//
//    public Map<String, BigInteger> getMaxStepLimits() throws Exception {
//        Map<String, BigInteger> map = new HashMap<>();
//        String[] types = {"invoke", "query"};
//        for(String t : types) {
//            RpcObject params = new RpcObject.Builder()
//                    .put("contextType", new RpcValue(t))
//                    .build();
//            BigInteger stepLimit = this.chainScore.call("getMaxStepLimit", params).asInteger();
//            map.put(t, stepLimit);
//        }
//        return map;
//    }
//
//    public void setMaxStepLimits(Map<String, BigInteger> limits)
//            throws ResultTimeoutException, TransactionFailureException, IOException {
//        List<Bytes> list = new LinkedList<>();
//        for(String type : limits.keySet()) {
//            RpcObject params = new RpcObject.Builder()
//                    .put("contextType", new RpcValue(type))
//                    .put("limit", new RpcValue(limits.get(type)))
//                    .build();
//            Bytes txHash = invoke(getWallet(), "setMaxStepLimit", params);
//            list.add(txHash);
//        }
//        for(Bytes txHash : list) {
//            TransactionResult result = getResult(txHash);
//            if (!Constants.STATUS_SUCCESS.equals(result.getStatus())) {
//                throw new TransactionFailureException(result.getFailure());
//            }
//        }
//    }
//
//    public Fee getFee() throws Exception {
//        Fee fee = new Fee();
//        fee.stepCosts = getStepCosts();
//        fee.stepMaxLimits = getMaxStepLimits();
//        fee.stepPrice = this.chainScore.getStepPrice();
//        return fee;
//    }
//
//    public void setFee(Fee fee) throws Exception {
//        setStepPrice(fee.stepPrice);
//        setStepCosts(fee.stepCosts);
//        setMaxStepLimits(fee.stepMaxLimits);
//    }
//
//    public TransactionResult addDeployer(Address address) throws IOException, ResultTimeoutException {
//        RpcObject params = new RpcObject.Builder()
//                .put("address", new RpcValue(address))
//                .build();
//        return invokeAndWaitResult(getWallet(), "addDeployer", params);
//    }
//
//    public TransactionResult removeDeployer(Address address) throws IOException, ResultTimeoutException {
//        RpcObject params = new RpcObject.Builder()
//                .put("address", new RpcValue(address))
//                .build();
//        return invokeAndWaitResult(getWallet(), "removeDeployer", params);
//    }
//
//    public TransactionResult setDeployerWhiteListEnabled(boolean yn) throws IOException, ResultTimeoutException {
//        RpcObject params = new RpcObject.Builder()
//                .put("yn", new RpcValue(yn))
//                .build();
//        return invokeAndWaitResult(getWallet(), "setDeployerWhiteListEnabled", params);
//    }

    // APIs for havah
    public TransactionResult setUSDTPrice(Wallet wallet, BigInteger price) throws IOException, ResultTimeoutException {
        RpcObject params = new RpcObject.Builder()
                .put("price", new RpcValue(price))
                .build();
        return invokeAndWaitResult(wallet, "setUSDTPrice", params);
    }

//    public int getUSDTPrice() throws IOException {
//        return call("getUSDTPrice", null).asInteger().intValue();
//    }

    public TransactionResult addPlanetManager(Wallet wallet, Address address) throws IOException, ResultTimeoutException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();
        return invokeAndWaitResult(wallet, "addPlanetManager", params);
    }

    public TransactionResult removePlanetManager(Wallet wallet, Address address) throws IOException, ResultTimeoutException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();
        return invokeAndWaitResult(wallet, "removePlanetManager", params);
    }

//    public TransactionResult claimPlanetReward(BigInteger[] ids) throws IOException, ResultTimeoutException {
//        RpcObject params = new RpcObject.Builder()
//                .put("ids", new RpcValue(ids))
//                .build();
//        return invokeAndWaitResult(getWallet(), "claimPlanetReward", params);
//    }

    public TransactionResult startRewardIssue(Wallet wallet, BigInteger height) throws IOException, ResultTimeoutException {
        RpcObject params = new RpcObject.Builder()
                .put("height", new RpcValue(height))
                .build();
        return invokeAndWaitResult(wallet, "startRewardIssue", params);
    }

    public TransactionResult setPrivateClaimableRate(Wallet wallet, BigInteger denominator, BigInteger numerator) throws IOException, ResultTimeoutException {
        RpcObject params = new RpcObject.Builder()
                .put("denominator", new RpcValue(denominator))
                .put("numerator", new RpcValue(numerator))
                .build();
        return invokeAndWaitResult(wallet, "setPrivateClaimableRate", params);
    }
}
