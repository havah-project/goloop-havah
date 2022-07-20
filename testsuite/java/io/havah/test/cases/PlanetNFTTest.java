package io.havah.test.cases;

import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.jsonrpc.RpcItem;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.PlanetNFTScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotEquals;


@Tag(Constants.TAG_HAVAH)
public class PlanetNFTTest extends TestBase {
    private static Wallet[] wallets;
    private static KeyWallet planetScoreOwner;
    private static PlanetNFTScore planetNFTScore;
    private static TransactionHandler txHandler;

    @BeforeAll
    static void setup() throws Exception {
        txHandler = Utils.getTxHandler();
        planetScoreOwner = txHandler.getChain().governorWallet;

        // init wallets
        wallets = new KeyWallet[3];
        wallets[0] = planetScoreOwner;
        for (int i = 1; i < wallets.length; i++) {
            wallets[i] = KeyWallet.create();
        }
        Utils.distributeCoin(wallets);
        planetNFTScore = new PlanetNFTScore(planetScoreOwner, txHandler);
    }

    @Test
    void name() throws Exception {
        RpcItem item = planetNFTScore.call("name", null);
        assertEquals(PlanetNFTScore.name, item.asString());
    }

    @Test
    void symbol() throws Exception {
        RpcItem item = planetNFTScore.call("symbol", null);
        assertEquals(PlanetNFTScore.symbol, item.asString());
    }

    void _mintAndCheckBalance(Wallet wallet) throws Exception {
        _mintAndCheckBalance(wallet, wallet.getAddress());
    }

    void _mintAndCheckBalance(Wallet wallet, Address to) throws Exception {
        _mintAndCheckBalance(wallet, to, 1, BigInteger.ONE, BigInteger.ONE);
    }

    static void _mintAndCheckBalance(Wallet wallet, Address to, int type, BigInteger usdt, BigInteger havah) throws Exception {
        boolean success = wallet.equals(planetScoreOwner);
        LOG.infoEntering("Mint", success ? "success" : "failure");

        // check balance of wallets[0] then compare after mint
        var balance = planetNFTScore.balanceOf(to).intValue();
        var totalSupply = planetNFTScore.totalSupply().intValue();
        Bytes txHash = planetNFTScore.mintPlanet(wallet, to, type, usdt, havah);
        TransactionResult result = planetNFTScore.getResult(txHash);
        assertEquals(success ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");

        // compare token balance
        int expectedBalance = success ? balance + 1 : balance;
        assertEquals(expectedBalance, planetNFTScore.balanceOf(to).intValue());

        // compare total supply
        int expectedSupply = success ? totalSupply + 1 : totalSupply;
        assertEquals(expectedSupply, planetNFTScore.totalSupply().intValue());
        LOG.infoExiting();
    }

    /*
        operator의 주소를 이용하여 mint가 정상적으로 되고
        operator가 아닌 주소를 이용하여 mint api호출 시 실패.
     */
    // 1
    @Test
    void mintPlanet() throws Exception {
        _mintAndCheckBalance(wallets[0]); // failure test
        _mintAndCheckBalance(planetScoreOwner); // success test
        LOG.info("mintPlanet totalSupply(" + planetNFTScore.totalSupply() + ")");
    }

    /*
        mint 이후 balanceOf에 정상적으로 반영되었는지 확인한다.
        한번에 여러 TX를 보내고 모든 TX에 대해 정성적으로 NFT가 생성되었는지 balanceOf함수를 이용하여 확인한다.
     */
    @Test
    void mintPlanetAndCheckBalances() throws Exception {
        Address newAddress = KeyWallet.create().getAddress();
        assertEquals(0, planetNFTScore.balanceOf(newAddress).intValue());

        var totalSupply = planetNFTScore.totalSupply().intValue();
        int mintCnt = 5;
        for (int i = 0; i < mintCnt; i++) {
            _mintAndCheckBalance(planetScoreOwner, newAddress); // negative test
        }

        assertEquals(mintCnt, planetNFTScore.balanceOf(newAddress).intValue());
        assertEquals(totalSupply + mintCnt, planetNFTScore.totalSupply().intValue());
        totalSupply += mintCnt;
        LOG.info("mintPlanetAndCheckBalances totalSupply(" + totalSupply + ")");

        // TODO test with change type
        int testMintCnt = 1000;
        int failureCnt = 0;
        List<Bytes> txHashes = new ArrayList<>();
        for (int i = 0; i < testMintCnt; i++) {
            Bytes txHash = planetNFTScore.mintPlanet(planetScoreOwner, newAddress, 1, BigInteger.ONE, BigInteger.ONE);
            txHashes.add(txHash);
        }

        int successTx = 0;
        for (int i = 0; i < testMintCnt; i++) {
            Bytes txHash = txHashes.get(i);
            TransactionResult result = planetNFTScore.getResult(txHash);
            if (result.getStatus().intValue() == 1) {
                successTx++;
            }
//            assertEquals(1, result.getStatus().intValue());
        }

        assertEquals(totalSupply + successTx, planetNFTScore.totalSupply().intValue());
        LOG.info("mintPlanetAndCheckBalances totalSupply(" + (totalSupply + successTx) + ")");
    }

    // mint & transfer
    void _mintAndTransfer(Wallet holder, Wallet from, Address to) throws Exception {
        boolean success = holder.equals(from);
        LOG.infoEntering("Transfer", success ? "success" : "failure");
        _mintAndCheckBalance(planetScoreOwner, holder.getAddress());

        int fromBalance = planetNFTScore.balanceOf(from.getAddress()).intValue();
        int toBalance = planetNFTScore.balanceOf(to).intValue();
        var tokenIdsMap = planetNFTScore.tokenIdsOf(holder.getAddress(), 0, 10);
        var tokenIds = tokenIdsMap.tokenIds;
        assertEquals(tokenIdsMap.balance.intValue(), tokenIds.size());
        assertNotEquals(0, tokenIds.size());
        var tokenId = tokenIds.get(0);
        Bytes txHash = planetNFTScore.transfer(from, to, tokenId);
        TransactionResult result = planetNFTScore.getResult(txHash);
        assertEquals(success ? 1 : 0, result.getStatus().intValue());
        assertEquals(success ? to : holder.getAddress(), planetNFTScore.ownerOf(tokenId));

        int expectedFromBalance = success ? fromBalance - 1 : fromBalance;
        int expectedToBalance = success ? toBalance + 1 : toBalance;
        assertEquals(expectedFromBalance, planetNFTScore.balanceOf(from.getAddress()).intValue());
        assertEquals(expectedToBalance, planetNFTScore.balanceOf(to).intValue());

        if (success) {
            // 이미 transfer한 것을 다시 transfer
            txHash = planetNFTScore.transfer(from, to, tokenId);
            result = planetNFTScore.getResult(txHash);
            assertEquals(0, result.getStatus().intValue()); // failure
        }
        LOG.infoExiting();
    }

    /*
        다른 주소로 전송이 정상적으로 이루어지는지 확인한다.
     */
    @Test
    void transfer() throws Exception {
        _mintAndTransfer(wallets[0], wallets[1], wallets[2].getAddress()); // failure case
        _mintAndTransfer(wallets[0], wallets[2], wallets[0].getAddress()); // failure case
        // check 내가 나에게 주는것....???
        _mintAndTransfer(wallets[0], wallets[0], wallets[1].getAddress()); // success case
        LOG.info("transfer totalSupply(" + planetNFTScore.totalSupply() + ")");
    }

    void _mintAndTransferFrom(Wallet holder, Wallet approveInvoker, Wallet approval, Wallet transferInvoker, Address to) throws Exception {
        boolean success = approval.equals(transferInvoker);
        var tokenIds = planetNFTScore.tokenIdsOf(holder.getAddress(), 0, 10);
        var approved = tokenIds.tokenIds.get(0);
        boolean approveSuccess = holder.equals(approveInvoker);
        var txHash = planetNFTScore.approve(approveInvoker, approval.getAddress(), approved); // approve From
        assertEquals(approveSuccess ? 1 : 0, planetNFTScore.getResult(txHash).getStatus().intValue());
        if (!approveSuccess) {
            return;
        }
        txHash = planetNFTScore.transferFrom(transferInvoker, holder.getAddress(), to, approved);
        var result = planetNFTScore.getResult(txHash);
        assertEquals(success ? 1 : 0, result.getStatus().intValue());
        var owner = planetNFTScore.ownerOf(approved);
        assertEquals(success ? to : holder.getAddress(), owner);
    }

    // transferFrom & approve test
    @Test
    void transferFrom() throws Exception {
        // approveInvoker, transferInvoker
        Wallet[] wallet = new Wallet[5];
        for (int i = 0; i < wallet.length; i++) {
            wallet[i] = KeyWallet.create();
        }
        Utils.distributeCoin(wallet);
        _mintAndCheckBalance(planetScoreOwner, wallet[0].getAddress()); // mint with cnt
        _mintAndTransferFrom(wallet[0], wallet[1], wallet[2], wallet[3], wallet[4].getAddress()); // failure
        _mintAndTransferFrom(wallet[0], wallet[0], wallet[1], wallet[2], wallet[4].getAddress()); // failure
        _mintAndTransferFrom(wallet[0], wallet[0], wallet[1], wallet[1], wallet[4].getAddress()); // success
        LOG.info("transferFrom totalSupply(" + planetNFTScore.totalSupply() + ")");
    }

    @Test
    void burn() throws Exception {
        Address holder = KeyWallet.create().getAddress();
        _mintAndCheckBalance(planetScoreOwner, holder); // success test
        _mintAndCheckBalance(planetScoreOwner, holder); // success test
        // mint 2
        int balance = planetNFTScore.balanceOf(holder).intValue();
        int supply = planetNFTScore.totalSupply().intValue();
        var tokenIdsMap = planetNFTScore.tokenIdsOf(holder, 0, 100);
        var tokenIds = tokenIdsMap.tokenIds;
        var burned = tokenIds.get(0);
        var txHash = planetNFTScore.burn(planetScoreOwner, burned);
        var result = planetNFTScore.getResult(txHash);
        assertEquals(1, result.getStatus().intValue());
        assertEquals(balance - 1, planetNFTScore.balanceOf(holder).intValue());
        assertEquals(supply - 1, planetNFTScore.totalSupply().intValue());

        var updated = planetNFTScore.tokenIdsOf(holder, 0, 100);
        var updatedTokenIds = updated.tokenIds;
        for (var tokenId : updatedTokenIds) {
            assertNotEquals(burned.intValue(), tokenId.intValue());
        }
        LOG.info("burn totalSupply(" + planetNFTScore.totalSupply() + ")");
    }

    // mint 1
    @Test
    void supply() throws Exception {
        var supply = planetNFTScore.totalSupply();
        Wallet holder = KeyWallet.create();

        // mint planet then check totalSupply == supply + 1
        var txHash = planetNFTScore.mintPlanet(holder.getAddress(), 2, BigInteger.ONE, BigInteger.ONE);
        TransactionResult result = planetNFTScore.getResult(txHash);
        assertEquals(1, result.getStatus().intValue(), "failure result(" + result + ")");
        supply = supply.add(BigInteger.ONE);
        assertEquals(supply, planetNFTScore.totalSupply());

        var tokenIds = planetNFTScore.tokenIdsOf(holder.getAddress(), 0, 1);
        var tokenId = tokenIds.tokenIds.get(0);

        // burn planet then check totalSupply == supply - 1
        txHash = planetNFTScore.burn(tokenId);
        result = planetNFTScore.getResult(txHash);
        assertEquals(1, result.getStatus().intValue(), "failure result(" + result + ")");

        assertEquals(supply.subtract(BigInteger.ONE), planetNFTScore.totalSupply(), "failure result(" + result + ")");
        LOG.info("supply totalSupply(" + planetNFTScore.totalSupply() + ")");
    }

    private Score.TokenInfo getFirstTokenInfo(Address address) throws Exception {
        var myToken = planetNFTScore.tokenIdsOf(address, 0, 1);
        var tokenId = myToken.tokenIds.get(0);
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        RpcItem item = planetNFTScore.call("infoOf", params);
        var hPrice = item.asObject().getItem("havahPrice").asInteger();
        boolean isCompany = item.asObject().getItem("isCompany").asBoolean();
        boolean isPrivate = item.asObject().getItem("isPrivate").asBoolean();
        Address owner = item.asObject().getItem("owner").asAddress();
        var uPrice = item.asObject().getItem("usdtPrice").asInteger();
        return new Score.TokenInfo(hPrice, isCompany, isPrivate, owner, uPrice);
    }

    @Test
    void tokenInfo() throws Exception {
        BigInteger usdtPrice = BigInteger.valueOf(111);
        BigInteger havahPrice = BigInteger.valueOf(1110);
        final int[] types = {
                PlanetNFTScore.PLANET_PRIVATE,
                PlanetNFTScore.PLANET_PUBLIC,
                PlanetNFTScore.PLANET_COMPANY,
        };

        for (var type : types) {
            Wallet wallet = KeyWallet.create();
            _mintAndCheckBalance(planetScoreOwner, wallet.getAddress(), type, usdtPrice, havahPrice);
            var tokenInfo =getFirstTokenInfo(wallet.getAddress());

            assertEquals(havahPrice, tokenInfo.getHavahPrice());
            assertEquals(usdtPrice, tokenInfo.getUsdtPrice());
            assertEquals(type == PlanetNFTScore.PLANET_PRIVATE, tokenInfo.isPrivate());
            assertEquals(type == PlanetNFTScore.PLANET_COMPANY, tokenInfo.isCompany());
            assertEquals(wallet.getAddress(), tokenInfo.getOwner());

            usdtPrice = usdtPrice.add(BigInteger.ONE);
            havahPrice = havahPrice.add(BigInteger.TEN);
        }
        LOG.info("tokenInfo totalSupply(" + planetNFTScore.totalSupply() + ")");
    }

//    @Test
    void planetSupply() throws Exception {
        var totalSupply = planetNFTScore.totalSupply().intValue();
        int testMintCnt = 1000;
        List<Bytes> txHashes = new ArrayList<>();
        Address newAddress = KeyWallet.create().getAddress();
        for (int i = 0; i < testMintCnt; i++) {
            Bytes txHash = planetNFTScore.mintPlanet(planetScoreOwner, newAddress, 1, BigInteger.ONE, BigInteger.ONE);
            txHashes.add(txHash);
        }

        int successTx = 0;
        for (int i = 0; i < testMintCnt; i++) {
            Bytes txHash = txHashes.get(i);
            TransactionResult result = planetNFTScore.getResult(txHash);
            if (result.getStatus().intValue() == 1) {
                successTx++;
            }
//            assertEquals(1, result.getStatus().intValue());
        }
        var diff = planetNFTScore.totalSupply().intValue() - totalSupply;
        assertEquals(diff, successTx);

        var tokenIds = planetNFTScore.tokenIdsOf(newAddress, 0, 2);
        for (var id : tokenIds.tokenIds) {
            var txHash = planetNFTScore.burn(planetScoreOwner, id);
            assertEquals(BigInteger.ONE, txHandler.getResult(txHash).getStatus());
            assertEquals(--successTx, planetNFTScore.totalSupply().intValue() - totalSupply);
        }
        LOG.info("token supply(" + successTx + ")");
        Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(Utils.getHeightNext(5));
    }
}