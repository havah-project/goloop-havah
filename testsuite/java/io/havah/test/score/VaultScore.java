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
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;

import java.io.IOException;
import java.math.BigInteger;
import java.util.Arrays;
import java.util.Map;

public class VaultScore extends Score {
    public static class VestingAccount {
        public Address account;
        public BigInteger amount; // 전체 분배 비율

        public VestingAccount(Address account, BigInteger amount) {
            this.account = account;
            this.amount = amount;
        }
    }

    public static class VestingHeight {
        public BigInteger height;
        public BigInteger ratio; // 텀 분배 비율

        public VestingHeight(BigInteger height, BigInteger ratio) {
            this.height = height;
            this.ratio = ratio;
        }
    }
    public VaultScore(TransactionHandler txHandler) {
        super(txHandler, Constants.VAULT_ADDRESS);
    }

    @Override
    public TransactionResult invokeAndWaitResult(Wallet wallet, String method, RpcObject params)
            throws ResultTimeoutException, IOException {
        return super.invokeAndWaitResult(wallet, method, params, BigInteger.ZERO, Constants.DEFAULT_STEPS);
    }

    public Bytes setAdmin(Wallet wallet, Address admin) throws Exception {
        RpcObject param = new RpcObject.Builder()
                .put("_admin", new RpcValue(admin))
                .build();
        return invoke(wallet, "setAdmin", param);
    }

    public Address admin() throws IOException {
        RpcObject params = new RpcObject.Builder()
                .build();
        return call("admin", params).asAddress();
    }

    public TransactionResult addAllocation(Wallet wallet, VestingAccount[] vestingAccounts)
            throws ResultTimeoutException, IOException {
        var accounts = new RpcArray.Builder();
        for (var a : vestingAccounts) {
            accounts.add(new RpcObject.Builder()
                    .put("account", new RpcValue(a.account))
                    .put("amount", new RpcValue(a.amount))
                    .build());
        }

        RpcObject params = new RpcObject.Builder()
                .put("vestingAccounts", accounts.build())
                .build();

        return invokeAndWaitResult(wallet, "addAllocation", params);
    }

    public TransactionResult setAllocation(Wallet wallet, VestingAccount vestingAccount)
            throws ResultTimeoutException, IOException {

        RpcObject params = new RpcObject.Builder()
                .put("account", new RpcObject.Builder()
                        .put("account", new RpcValue(vestingAccount.account))
                        .put("amount", new RpcValue(vestingAccount.amount))
                        .build())
                .build();

        return invokeAndWaitResult(wallet, "setAllocation", params);
    }

    public TransactionResult setVestingHeights(Wallet wallet, VestingHeight[] vestingHeights)
            throws ResultTimeoutException, IOException {
        var heights = new RpcArray.Builder();
        for (var h : vestingHeights) {
            heights.add(new RpcObject.Builder()
                    .put("height", new RpcValue(h.height))
                    .put("ratio", new RpcValue(h.ratio))
                    .build());
        }

        RpcObject params = new RpcObject.Builder()
                .put("vestingHeights", heights.build())
                .build();

        return invokeAndWaitResult(wallet, "setVestingHeights", params);
    }

    public TransactionResult claim(Wallet wallet)
            throws ResultTimeoutException, IOException {
        return invokeAndWaitResult(wallet, "claim", null);
    }

    public Map<String, Object> getVestingInfo(Address owner) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(owner))
                .build();
        RpcObject item = (RpcObject) call("getVestingInfo", params);
        if (!item.isEmpty()) {
            RpcArray heightsArr = item.getItem("heights").asArray();
            BigInteger[] heights = new BigInteger[heightsArr.size()];
            for (int i=0; i< heights.length; i++) {
                heights[i] = heightsArr.get(i).asInteger();
            }
            RpcArray amountsArr = item.getItem("vestingAmounts").asArray();
            BigInteger[] acmounts = new BigInteger[amountsArr.size()];
            for (int i=0; i< acmounts.length; i++) {
                acmounts[i] = amountsArr.get(i).asInteger();
            }
            return Map.of(
                    "total", item.getItem("total").asInteger(),
                    "heights", Arrays.toString(heights),
                    "claimable", item.getItem("claimable").asInteger(),
                    "vestingAmounts", Arrays.toString(acmounts)
            );
        }
        return Map.of();
    }
}
