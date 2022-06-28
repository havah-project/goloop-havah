package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;
import score.annotation.Optional;

import java.io.IOException;
import java.math.BigInteger;

public class SustainableFundScore extends Score {
    public static final String name = "HAVAH Planet";
    public static final String symbol = "HAPL";

    public SustainableFundScore(TransactionHandler txHandler) {
        super(txHandler, Constants.SUSTAINABLEFUND_ADDRESS);
    }

    public RpcObject getInflow() throws IOException {
        return call("getInflow", null).asObject();
    }

    public RpcObject getOutflow() throws IOException {
        return call("getOutflow", null).asObject();
    }

    public Bytes transfer(Wallet wallet, Address to, BigInteger value) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_to", new RpcValue(to))
                .put("_value", new RpcValue(value))
                .build();
        return invoke(wallet, "transfer", params);
    }

    public Bytes transferToken(Wallet wallet, Address token, Address to, BigInteger value) throws Exception {
        return transferToken(wallet, token, to, value, null);
    }

    public Bytes transferToken(Wallet wallet, Address token, Address to, BigInteger value, byte[] data) throws Exception {
        RpcObject.Builder builder = new RpcObject.Builder()
                .put("_tokenAddress", new RpcValue(token))
                .put("_to", new RpcValue(to))
                .put("_value", new RpcValue(value));
        if (data != null) {
            builder.put("_data", new RpcValue(data));
        }
        return invoke(wallet, "transferToken", builder.build());
    }

    public Bytes transferFromPlanetSales(Wallet wallet, BigInteger value) throws Exception {
        return invoke(wallet, "transferFromPlanetSales", null, value);
    }

    public Bytes transferFromServiceFee(Wallet wallet, BigInteger value) throws Exception {
        return invoke(wallet, "transferFromServiceFee", null, value);
    }
}
