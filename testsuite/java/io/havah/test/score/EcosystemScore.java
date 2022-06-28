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

    public RpcArray getLockupSchedule() throws Exception {
        return call("getLockupSchedule", null).asArray();
    }
}
