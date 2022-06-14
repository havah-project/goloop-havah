package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.transport.jsonrpc.RpcItem;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;

import java.io.IOException;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import static foundation.icon.test.common.Env.LOG;

public class PlanetNFTScore extends Score {
    public static final String name = "HAVAH Planet";
    public static final String symbol = "HAPL";

    public PlanetNFTScore(TransactionHandler txHandler) {
        super(txHandler, Constants.PLANETNFT_ADDRESS);
    }

    public Address ownerOf(BigInteger tokenId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return call("ownerOf", params).asAddress();
    }

    public Address getApproved(BigInteger tokenId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return call("getApproved", params).asAddress();
    }

    public BigInteger balanceOf(Address owner) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_owner", new RpcValue(owner))
                .build();
        return call("balanceOf", params).asInteger();
    }

    public BigInteger totalSupply() throws IOException {
        return call("totalSupply", null).asInteger();
    }

    public BigInteger tokenByIndex(int index) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_index", new RpcValue(BigInteger.valueOf(index)))
                .build();
        return call("tokenByIndex", params).asInteger();
    }

    public BigInteger tokenOfOwnerByIndex(Address owner, int index) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_owner", new RpcValue(owner))
                .put("_index", new RpcValue(BigInteger.valueOf(index)))
                .build();
        return call("tokenOfOwnerByIndex", params).asInteger();
    }

    public Bytes mintPlanet(Wallet wallet, Address _to, int _type, BigInteger _priceInUSDT, BigInteger _priceInHVH) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_to", new RpcValue(_to))
                .put("_type", new RpcValue(BigInteger.valueOf(_type)))
                .put("_priceInUSDT", new RpcValue(_priceInUSDT))
                .put("_priceInHVH", new RpcValue(_priceInHVH))
                .build();
        return invoke(wallet, "mintPlanet", params);
    }

    public Bytes burn(Wallet wallet, BigInteger tokenId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(wallet, "burn", params);
    }

    public Bytes approve(Wallet wallet, Address to, BigInteger tokenId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_to", new RpcValue(to))
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(wallet, "approve", params);
    }

    public Bytes transfer(Wallet wallet, Address to, BigInteger tokenId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_to", new RpcValue(to))
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(wallet, "transfer", params);
    }

    public Bytes transferFrom(Wallet wallet, Address from, Address to, BigInteger tokenId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_from", new RpcValue(from))
                .put("_to", new RpcValue(to))
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(wallet, "transferFrom", params);
    }

    public TokenIds tokenIdsOf(Address owner, int start, int count) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_owner", new RpcValue(owner))
                .put("_start", new RpcValue(BigInteger.valueOf(start)))
                .put("_count", new RpcValue(BigInteger.valueOf(count)))
                .build();
        var object = call("tokenIdsOf", params).asObject();
        List<BigInteger> ids = new ArrayList<>();
        var array = object.getItem("tokenIds").asArray();
        for (RpcItem rpcItem : array) {
            ids.add(rpcItem.asInteger());
        }
        return new TokenIds(ids, object.getItem("balance").asInteger());
    }

    public static class TokenIds {
        public final List<BigInteger> tokenIds;
        public final BigInteger balance;

        TokenIds(List<BigInteger> tokenIds, BigInteger balance) {
            this.tokenIds = tokenIds;
            this.balance = balance;
        }
    }

    public Bytes requestStartOp(Wallet wallet, BigInteger tokenId, Address agentAddr) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .put("_agentAddr", new RpcValue(agentAddr))
                .build();
        return invoke(wallet, "requestStartOp", params);
    }

    public Bytes cancelStartOpReq(Wallet wallet, BigInteger tokenId) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(wallet, "cancelStartOpReq", params);
    }

    public Bytes rejectStartOpReq(Wallet wallet, BigInteger tokenId) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(wallet, "rejectStartOpReq", params);
    }

    public Bytes acceptStartOpReq(Wallet wallet, BigInteger tokenId, Address receiver) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .put("_receiver", new RpcValue(receiver))
                .build();
        return invoke(wallet, "acceptStartOpReq", params);
    }

    public Bytes requestStopOp(Wallet wallet, BigInteger tokenId) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(wallet, "requestStopOp", params);
    }

    public Bytes cancelStopOpReq(Wallet wallet, BigInteger tokenId) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(wallet, "cancelStopOpReq", params);
    }

    public Bytes acceptStopOpReq(Wallet wallet, BigInteger tokenId) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(wallet, "acceptStopOpReq", params);
    }

    // agent apis
    // agentBalanceOf
    // agentOf
    // requestStartOp
    // cancelStartOpReq
    // rejectStartOpReq
    // acceptStartOpReq
    // requestStopOp
    // cancelStopOpReq
    // acceptStopOpReq
    // getOpState
    // lastTokenId
    // mintedTokenCnt
}
