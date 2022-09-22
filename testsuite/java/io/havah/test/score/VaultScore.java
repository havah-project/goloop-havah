package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.jsonrpc.RpcArray;
import foundation.icon.icx.transport.jsonrpc.RpcItem;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.ResultTimeoutException;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;

import java.io.IOException;
import java.math.BigInteger;

public class VaultScore extends Score {
    public static class VestingAccount {
        public Address address;
        public BigInteger amount; // 전체 분배 비율

        public VestingAccount(Address address, BigInteger amount) {
            this.address = address;
            this.amount = amount;
        }

        public String toString() {
            StringBuilder builder = new StringBuilder();
            builder.append("{\"");
            builder.append(address.toString());
            builder.append("\",\"");
            builder.append(amount.toString());
            builder.append("\"}");
            return builder.toString();
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

        public String toString() {
            StringBuilder builder = new StringBuilder();
            builder.append("{\"");
            builder.append(timestamp.toString());
            builder.append("\",\"");
            builder.append(numerator.toString());
            builder.append("\",\"");
            builder.append(denominator.toString());
            builder.append("\"}");
            return builder.toString();
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

    public TransactionResult addAllocation(Wallet wallet, VestingAccount[] vestingAccounts) throws ResultTimeoutException, IOException {
        return addAllocation(wallet, vestingAccounts, Constants.DEFAULT_STEPS);
    }

    public TransactionResult addAllocation(Wallet wallet, VestingAccount[] vestingAccounts, BigInteger steps)
            throws ResultTimeoutException, IOException {
        var accounts = new RpcArray.Builder();
        for (var a : vestingAccounts) {
            accounts.add(new RpcObject.Builder()
                    .put("address", new RpcValue(a.address))
                    .put("amount", new RpcValue(a.amount))
                    .build());
        }

        RpcObject params = new RpcObject.Builder()
                .put("_vestingAccounts", accounts.build())
                .build();

        return invokeAndWaitResult(wallet, "addAllocation", params, BigInteger.ZERO, steps);
    }

    public TransactionResult setAllocation(Wallet wallet, VestingAccount vestingAccount)
            throws ResultTimeoutException, IOException {

        RpcObject params = new RpcObject.Builder()
                .put("_account", new RpcObject.Builder()
                        .put("address", new RpcValue(vestingAccount.address))
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
                .put("_account", new RpcValue(account))
                .put("_vestingSchedules", heights.build())
                .build();

        return invokeAndWaitResult(wallet, "setVestingSchedules", params);
    }

    public TransactionResult claim(Wallet wallet)
            throws ResultTimeoutException, IOException {
        return invokeAndWaitResult(wallet, "claim", null);
    }

    public BigInteger getClaimable(Address address) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_address", new RpcValue(address))
                .build();

        return call("getClaimable", params).asInteger();
    }

    public BigInteger getAllocation(Address address) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_address", new RpcValue(address))
                .build();

        RpcItem val = call("getAllocation", params);
        return val == null ? null : val.asInteger();
    }

    public void ensureAddAllocation(TransactionResult result, String vestingAccounts)
            throws IOException {
        TransactionResult.EventLog event = findEventLog(result, "AddAllocation(str)");
        if (event != null) {
            String _vestingAccounts = event.getIndexed().get(1).asString();
            if (_vestingAccounts.equals(vestingAccounts)) {
                return; // ensured
            }
        }
        throw new IOException("ensureAddAllocation failed.");
    }

    public void ensureSetAllocation(TransactionResult result, String account)
            throws IOException {
        TransactionResult.EventLog event = findEventLog(result, "SetAllocation(str)");
        if (event != null) {
            String _accounts = event.getIndexed().get(1).asString();
            if (_accounts.equals(account)) {
                return; // ensured
            }
        }
        throw new IOException("ensureSetAllocation failed.");
    }

    public void ensureSetVestingSchedules(TransactionResult result, Address account, String vestingSchedules)
            throws IOException {
        TransactionResult.EventLog event = findEventLog(result, "SetVestingSchedules(Address,str)");
        if (event != null) {
            Address _accounts = event.getIndexed().get(1).asAddress();
            String _vestingSchedules = event.getIndexed().get(2).asString();
            if (_accounts.equals(account) && _vestingSchedules.equals(vestingSchedules)) {
                return; // ensured
            }
        }
        throw new IOException("ensureSetVestingSchedules failed.");
    }

    public void ensureClaim(TransactionResult result, Address account, BigInteger amount)
            throws IOException {
        TransactionResult.EventLog event = findEventLog(result, "Claim(Address,int)");
        if (event != null) {
            Address _accounts = event.getIndexed().get(1).asAddress();
            BigInteger _amount = event.getIndexed().get(2).asInteger();
            if (_accounts.equals(account) && _amount.equals(amount)) {
                return; // ensured
            }
        }
        throw new IOException("ensureClaim failed.");
    }
}
