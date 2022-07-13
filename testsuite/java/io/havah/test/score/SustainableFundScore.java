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

    public static final String INFLOW_TXFEE = "TX_FEE";
    public static final String INFLOW_PLANETSALES = "PLANET_SALES";
    public static final String INFLOW_MISSINGREWARD = "MISSING_REWARD";
    public static final String INFLOW_SERVICEFEE = "SERVICE_FEE";
    public static final String OUTFLOW_CUSTOM = "CUSTOM";


    public SustainableFundScore(TransactionHandler txHandler) {
        super(txHandler, Constants.SUSTAINABLEFUND_ADDRESS);
    }

    public RpcObject getInflow() throws IOException {
        return call("getInflow", null).asObject();
    }

    public RpcObject getInflowInUSDT() throws IOException {
        return call("getInflowInUSDT", null).asObject();
    }

    public BigInteger getInflowAmount() throws Exception {
        var inflow = getInflow();
        return inflow.getItem(INFLOW_TXFEE).asInteger()
                .add(inflow.getItem(INFLOW_MISSINGREWARD).asInteger())
                .add(inflow.getItem(INFLOW_PLANETSALES).asInteger())
                .add(inflow.getItem(INFLOW_SERVICEFEE).asInteger());
    }

    public RpcObject getOutflow() throws IOException {
        return call("getOutflow", null).asObject();
    }

    public RpcObject getOutflowInUSDT() throws IOException {
        return call("getOutflowInUSDT", null).asObject();
    }

    public Bytes setUsdt(Wallet wallet, Address address) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_address", new RpcValue(address))
                .build();
        return invoke(wallet, "setUsdt", params);
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

    public Bytes depositFromPlanetSales(Wallet wallet, BigInteger value) throws Exception {
        return invoke(wallet, "depositFromPlanetSales", null, value);
    }

    public Bytes depositFromServiceFee(Wallet wallet, BigInteger value) throws Exception {
        return invoke(wallet, "depositFromServiceFee", null, value);
    }
}
