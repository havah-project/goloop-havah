package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.jsonrpc.*;
import foundation.icon.test.common.ResultTimeoutException;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;

import java.io.IOException;
import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;

public class VaultScore extends Score {
    public static class VestingAccount {
        public Address account;
        public BigInteger scheduledPercents; // 분배 비율

        public VestingAccount(Address account, BigInteger percents) {
            this.account = account;
            this.scheduledPercents = percents;
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

    public TransactionResult setAccounts(Wallet wallet, VestingAccount[] vestingAccounts, BigInteger[] heights)
            throws ResultTimeoutException, IOException {
        var accounts = new RpcArray.Builder();
        for (var a : vestingAccounts) {
            accounts.add(new RpcObject.Builder()
                    .put("account", new RpcValue(a.account))
                    .put("scheduledPercents", new RpcValue(a.scheduledPercents))
                    .build());
        }

        var vestingHeights = new RpcArray.Builder();
        for (var h : heights) {
            vestingHeights.add(new RpcValue(h));
        }

        RpcObject params = new RpcObject.Builder()
                .put("vestingAccounts", accounts.build())
                .put("heights", vestingHeights.build())
                .build();
        return invokeAndWaitResult(wallet, "setAccounts", params);
    }

    public TransactionResult claim(Wallet wallet)
            throws ResultTimeoutException, IOException {
        return invokeAndWaitResult(wallet, "claim", null);
    }

    public BigInteger getBalanceOf(Address owner) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(owner))
                .build();
        return call("getBalanceOf", params).asInteger();
    }

    public BigInteger getClaimableAmount(Address owner) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(owner))
                .build();
        try {
            RpcItem item = call("getClaimableAmount", params);
            return item.asInteger();
        } catch (RpcError e) {
            LOG.info("getClaimableAmount rpc error = " + e.getMessage());
        }

        return BigInteger.ZERO;
    }
}
