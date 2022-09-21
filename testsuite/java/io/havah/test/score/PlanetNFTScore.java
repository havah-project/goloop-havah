package io.havah.test.score;

import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.transport.jsonrpc.RpcArray;
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

public class PlanetNFTScore extends Score {
    public static final String name = "HAVAH Planet";
    public static final String symbol = "HAPL";
    private final Wallet deployer;

    public static final int PLANET_PRIVATE = 1;
    public static final int PLANET_PUBLIC = 2;
    public static final int PLANET_COMPANY = 4;


    public PlanetNFTScore(Wallet deployer, TransactionHandler txHandler) {
        super(txHandler, Constants.PLANETNFT_ADDRESS);
        this.deployer = deployer;
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

    public BigInteger tokenByIndex(BigInteger index) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_index", new RpcValue(index))
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

    public static BigInteger serialTokenId = BigInteger.ZERO;

    public Bytes mintPlanet(Address _to, int _type, BigInteger _priceInUSDT, BigInteger _priceInHVH, BigInteger _tokenId) throws IOException {
        return mintPlanet(deployer, _to, _type, _priceInUSDT, _priceInHVH, _tokenId);
    }

    public Bytes mintPlanet(Address _to, int _type, BigInteger _priceInUSDT, BigInteger _priceInHVH) throws IOException {
        return mintPlanet(deployer, _to, _type, _priceInUSDT, _priceInHVH);
    }

    public Bytes mintPlanet(Wallet wallet, Address _to, int _type,
                            BigInteger _priceInUSDT, BigInteger _priceInHVH) throws IOException {
        serialTokenId = serialTokenId.add(BigInteger.ONE);
        return mintPlanet(wallet, _to, _type, _priceInUSDT, _priceInHVH, serialTokenId);
    }

    public Bytes mintPlanet(Wallet wallet, Address _to, int _type,
                            BigInteger _priceInUSDT, BigInteger _priceInHVH, BigInteger _tokenId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_priceInUSDT", new RpcValue(_priceInUSDT))
                .put("_priceInHVH", new RpcValue(_priceInHVH))
                .put("_tokenId", new RpcValue(_tokenId))
                .put("_to", new RpcValue(_to))
                .put("_type", new RpcValue(BigInteger.valueOf(_type)))
                .build();
        return invoke(wallet, "mintPlanet", params);
    }

    public Bytes mintPlanet(Wallet wallet, List<MintInfo> _info, Address _to, int _type) throws IOException {
        var mintInfo = new RpcArray.Builder();
        for (var param : _info) {
            mintInfo.add(new RpcObject.Builder()
                    .put("_priceInUSDT", new RpcValue(param.getPriceInUSDT()))
                    .put("_priceInHVH", new RpcValue(param.getPriceInHVH()))
                    .put("_tokenId", new RpcValue(param.getTokenId()))
                    .build());
        }

        RpcObject params = new RpcObject.Builder()
                .put("_info", mintInfo.build())
                .put("_to", new RpcValue(_to))
                .put("_type", new RpcValue(BigInteger.valueOf(_type)))
                .build();
        return invoke(wallet, "mintPlanetBatch", params);
    }

    public Bytes burn(BigInteger tokenId) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return invoke(deployer, "burn", params);
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

    public RpcObject infoOf(BigInteger tokenId) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        return call("infoOf", params).asObject();
    }

    public Bytes setTransferable(Wallet wallet, int type, boolean transferable) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_type", new RpcValue(BigInteger.valueOf(type)))
                .put("_transferable", new RpcValue(transferable))
                .build();
        return invoke(wallet, "setTransferable", params);
    }

    public boolean isTransferable(int type) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_type", new RpcValue(BigInteger.valueOf(type)))
                .build();
        var result = call("isTransferable", params).asBoolean();
        return result;
    }

    public Bytes setAdmin(Wallet wallet, Address admin) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_admin", new RpcValue(admin))
                .build();
        return invoke(wallet, "setAdmin", params);
    }

    public Bytes addMintingApprover(Wallet wallet, Address approver) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_address", new RpcValue(approver))
                .build();
        return invoke(wallet, "addMintingApprover", params);
    }

    public Bytes removeMintingApprover(Wallet wallet, Address approver) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_address", new RpcValue(approver))
                .build();
        return invoke(wallet, "removeMintingApprover", params);
    }

    public List<Address> getMintingApprover() throws IOException {
        var array = call("mintingApprover", null).asArray();
        List<Address> approvers = new ArrayList<>();
        for (RpcItem rpcItem : array) {
            approvers.add(rpcItem.asAddress());
        }
        return approvers;
    }

    public static class TokenIds {
        public final List<BigInteger> tokenIds;
        public final BigInteger balance;

        TokenIds(List<BigInteger> tokenIds, BigInteger balance) {
            this.tokenIds = tokenIds;
            this.balance = balance;
        }
    }

    public static TokenInfo toTokenInfo(RpcObject object) {
        return new TokenInfo(object.getItem("havahPrice").asInteger(), object.getItem("isCompany").asBoolean(),
                object.getItem("isPrivate").asBoolean(), object.getItem("owner").asAddress(),
                object.getItem("usdtPrice").asInteger(), object.getItem("height").asInteger());
    }

    public static class TokenInfo {
        private final BigInteger havahPrice;
        private final boolean isCompany;
        private final boolean isPrivate;
        private final Address owner;
        private final BigInteger usdtPrice;
        private final BigInteger height;

        public TokenInfo(BigInteger havahPrice, boolean isCompany, boolean isPrivate, Address owner, BigInteger usdtPrice, BigInteger height) {
            this.havahPrice = havahPrice;
            this.isCompany = isCompany;
            this.isPrivate = isPrivate;
            this.owner = owner;
            this.usdtPrice = usdtPrice;
            this.height = height;
        }

        public BigInteger getHavahPrice() {
            return havahPrice;
        }

        public boolean isCompany() {
            return isCompany;
        }

        public boolean isPrivate() {
            return isPrivate;
        }

        public Address getOwner() {
            return owner;
        }

        public BigInteger getUsdtPrice() {
            return usdtPrice;
        }

        public BigInteger getHeight() {
            return height;
        }
    }

    public static class MintInfo {
        private final BigInteger tokenId;
        private final BigInteger priceInUSDT;
        private final BigInteger priceInHVH;

        public MintInfo(BigInteger tokenId, BigInteger priceInUSDT, BigInteger priceInHVH) {
            this.tokenId = tokenId;
            this.priceInUSDT = priceInUSDT;
            this.priceInHVH = priceInHVH;
        }

        public BigInteger getTokenId() {
            return tokenId;
        }

        public BigInteger getPriceInUSDT() {
            return priceInUSDT;
        }

        public BigInteger getPriceInHVH() {
            return priceInHVH;
        }
    }

    public static MintInfo mintInfo(BigInteger tokenId, BigInteger priceInUSDT, BigInteger priceInHVH) {
        return new MintInfo(tokenId, priceInUSDT, priceInHVH);
    }
}
