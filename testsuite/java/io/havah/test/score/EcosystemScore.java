package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.transport.jsonrpc.RpcArray;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import static foundation.icon.test.common.Env.LOG;

public class EcosystemScore extends Score {

    public EcosystemScore(TransactionHandler txHandler) {
        super(txHandler, Constants.ECOSYSTEM_ADDRESS);
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
            var height = s.asObject().getItem("BLOCK_HEIGHT").asInteger();
            var amount = s.asObject().getItem("LOCKUP_AMOUNT").asInteger();
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
        private final BigInteger blockHeight;
        private final BigInteger amount;
        public LockupSchedule(BigInteger blockHeight, BigInteger amount) {
            this.blockHeight = blockHeight;
            this.amount = amount;
        }

        public BigInteger getBlockHeight() {
            return blockHeight;
        }

        public BigInteger getAmount() {
            return amount;
        }

        @Override
        public String toString() {
            return "LockSchedule{" +
                    "blockHeight=" + blockHeight +
                    ", amount=" + amount +
                    '}';
        }
    }
}
