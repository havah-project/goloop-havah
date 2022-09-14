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

    public static class VestingSchedule {
        public BigInteger timestamp;
        public BigInteger denominator; // 분모
        public BigInteger numerator; // 분자

        public VestingSchedule(BigInteger _timestamp, BigInteger _denominator, BigInteger _numerator) {
            this.timestamp = _timestamp;
            this.denominator = _denominator;
            this.numerator = _numerator;
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

    public TransactionResult setVestingSchedules(Wallet wallet, Address account, VestingSchedule[] vestingSchedules)
            throws ResultTimeoutException, IOException {
        var heights = new RpcArray.Builder();
        for (var v : vestingSchedules) {
            heights.add(new RpcObject.Builder()
                    .put("timestamp", new RpcValue(v.timestamp))
                    .put("denominator", new RpcValue(v.denominator))
                    .put("numerator", new RpcValue(v.numerator))
                    .build());
        }

        RpcObject params = new RpcObject.Builder()
                .put("account", new RpcValue(account))
                .put("vestingSchedules", heights.build())
                .build();

        return invokeAndWaitResult(wallet, "setVestingSchedules", params);
    }

    public TransactionResult claim(Wallet wallet)
            throws ResultTimeoutException, IOException {
        return invokeAndWaitResult(wallet, "claim", null);
    }

    public BigInteger getClaimable(Address address) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();

        return call("getClaimable", params).asInteger();
    }

    public BigInteger getAllocation(Address address) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("address", new RpcValue(address))
                .build();

        return call("getAllocation", params).asInteger();
    }
}
