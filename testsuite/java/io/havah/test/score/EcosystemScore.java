package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

public class EcosystemScore extends Score {

    public EcosystemScore(TransactionHandler txHandler) {
        super(txHandler, Constants.ECOSYSTEM_ADDRESS);
    }

    public EcosystemScore(Score score) {
        super(score);
    }

    public Bytes transfer(Wallet wallet, Address to, BigInteger amount) throws Exception {
        RpcObject param = new RpcObject.Builder()
                .put("_to", new RpcValue(to))
                .put("_value", new RpcValue(amount))
                .build();
        return invoke(wallet, "transfer", param);
    }

    public List<LockupSchedule> getLockupSchedule() throws Exception {
        var schedules =  call("getLockupSchedule", null).asArray();
        List<LockupSchedule> list = new ArrayList<>();
        for (var s : schedules) {
            var height = s.asObject().getItem("timestamp").asInteger();
            var amount = s.asObject().getItem("lockup_amount").asInteger();
            list.add(new LockupSchedule(height, amount));
        }
        return list;
    }

    public Bytes setAdmin(Wallet wallet, Address admin) throws Exception {
        RpcObject param = new RpcObject.Builder()
                .put("_admin", new RpcValue(admin))
                .build();
        return invoke(wallet, "setAdmin", param);
    }

    public static class LockupSchedule {
        private final BigInteger timestamp;
        private final BigInteger amount;
        public LockupSchedule(BigInteger blockTimestamp, BigInteger amount) {
            this.timestamp = blockTimestamp;
            this.amount = amount;
        }

        public BigInteger getTimestamp() {
            return timestamp;
        }

        public BigInteger getAmount() {
            return amount;
        }

        @Override
        public String toString() {
            return "LockSchedule{" +
                    "blockTimestamp=" + timestamp +
                    ", amount=" + amount +
                    '}';
        }
    }
}
