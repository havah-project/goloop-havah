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
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.PlanetNFTScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.*;


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
        _mintAndCheckBalance(wallet, to, PlanetNFTScore.PLANET_PUBLIC, BigInteger.ONE, BigInteger.ONE);
    }

    void _mintAndCheckBalance(Wallet wallet, Address to, boolean expected) throws Exception {
        _mintAndCheckBalance(wallet, to, PlanetNFTScore.PLANET_PUBLIC, BigInteger.ONE, BigInteger.ONE, expected);
    }

    static void _mintAndCheckBalance(Wallet wallet, Address to, int type, BigInteger usdt, BigInteger havah) throws Exception {
        _mintAndCheckBalance(wallet, to, type, usdt, havah, wallet.equals(planetScoreOwner));
    }

    static void _mintAndCheckBalance(Wallet wallet, Address to, int type, BigInteger usdt, BigInteger havah, boolean expected) throws Exception {
        LOG.infoEntering("Mint", expected ? "success" : "failure");

        // check balance of wallets[0] then compare after mint
        var balance = planetNFTScore.balanceOf(to).intValue();
        var totalSupply = planetNFTScore.totalSupply().intValue();
        Bytes txHash = planetNFTScore.mintPlanet(wallet, to, type, usdt, havah);
        TransactionResult result = planetNFTScore.getResult(txHash);
        assertEquals(expected ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");

        // compare token balance
        int expectedBalance = expected ? balance + 1 : balance;
        assertEquals(expectedBalance, planetNFTScore.balanceOf(to).intValue());

        // compare total supply
        int expectedSupply = expected ? totalSupply + 1 : totalSupply;
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
        }

        assertEquals(totalSupply + successTx, planetNFTScore.totalSupply().intValue());
        LOG.info("mintPlanetAndCheckBalances totalSupply(" + (totalSupply + successTx) + ")");
    }

    // mint & transfer
    void _mintAndTransfer(Wallet holder, Wallet from, Address to) throws Exception {
        boolean success = holder.equals(from) && planetNFTScore.isTransferable(PlanetNFTScore.PLANET_PUBLIC);
        LOG.infoEntering("Transfer", success ? "success" : "failure");
        _mintAndCheckBalance(planetScoreOwner, holder.getAddress(), PlanetNFTScore.PLANET_PUBLIC, BigInteger.ONE, BigInteger.ONE);

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
        boolean isTransferable = false;
        for (int i = 0; i < 2; i++) {
            System.out.println("transfer : isTransferable(" + isTransferable + ")");
            assertEquals(isTransferable, planetNFTScore.isTransferable(PlanetNFTScore.PLANET_PUBLIC));
            _mintAndTransfer(wallets[0], wallets[1], wallets[2].getAddress()); // failure case
            _mintAndTransfer(wallets[0], wallets[2], wallets[0].getAddress()); // failure case
            _mintAndTransfer(wallets[0], wallets[0], wallets[1].getAddress()); // success case
            isTransferable = !isTransferable;
            System.out.println("transfer : setTransferable(" + isTransferable + ")");
            var txHash = planetNFTScore.setTransferable(planetScoreOwner, PlanetNFTScore.PLANET_PUBLIC, isTransferable);
            assertSuccess(txHash);
        }
        LOG.info("transfer totalSupply(" + planetNFTScore.totalSupply() + ")");
    }

    void _mintAndTransferFrom(Wallet holder, Wallet approveInvoker, Wallet approver, Wallet transferInvoker, Address to) throws Exception {
        boolean success = approver.equals(transferInvoker) && planetNFTScore.isTransferable(PlanetNFTScore.PLANET_PUBLIC);
        var tokenIds = planetNFTScore.tokenIdsOf(holder.getAddress(), 0, 10);
        var approved = tokenIds.tokenIds.get(0);
        boolean approveSuccess = holder.equals(approveInvoker);
        var txHash = planetNFTScore.approve(approveInvoker, approver.getAddress(), approved); // approve From
        assertEquals(approveSuccess ? 1 : 0, planetNFTScore.getResult(txHash).getStatus().intValue());
        if (!approveSuccess) {
            return;
        }
        txHash = planetNFTScore.transferFrom(transferInvoker, holder.getAddress(), to, approved);
        var result = txHandler.getResult(txHash);
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
        boolean isTransferable = false;
        for (int i = 0; i < 2; i++) {
            System.out.println("transferFrom : isTransferable(" + isTransferable + ")");
            assertEquals(isTransferable, planetNFTScore.isTransferable(PlanetNFTScore.PLANET_PUBLIC));
            _mintAndCheckBalance(planetScoreOwner, wallet[0].getAddress()); // mint with cnt
            _mintAndTransferFrom(wallet[0], wallet[1], wallet[2], wallet[3], wallet[4].getAddress()); // failure
            _mintAndTransferFrom(wallet[0], wallet[0], wallet[1], wallet[2], wallet[4].getAddress()); // failure
            _mintAndTransferFrom(wallet[0], wallet[0], wallet[1], wallet[1], wallet[4].getAddress()); // success
            isTransferable = !isTransferable;
            System.out.println("transferFrom : setTransferable(" + isTransferable + ")");
            var txHash = planetNFTScore.setTransferable(planetScoreOwner, PlanetNFTScore.PLANET_PUBLIC, isTransferable);
            assertSuccess(txHash);
        }

        LOG.info("transferFrom totalSupply(" + planetNFTScore.totalSupply() + ")");
    }

    void _burnAndCheckBalance(Wallet wallet, Address holder, boolean expected) throws Exception {
        int balance = planetNFTScore.balanceOf(holder).intValue();
        int supply = planetNFTScore.totalSupply().intValue();
        var tokenIdsMap = planetNFTScore.tokenIdsOf(holder, 0, 100);
        var tokenIds = tokenIdsMap.tokenIds;
        var burned = tokenIds.get(0);
        var txHash = planetNFTScore.burn(wallet, burned);
        var result = planetNFTScore.getResult(txHash);
        int sub = 0;
        if (expected) {
            assertSuccess(result);
            sub = 1;
        } else {
            assertFailure(result);
        }
        assertEquals(balance - sub, planetNFTScore.balanceOf(holder).intValue());
        assertEquals(supply - sub, planetNFTScore.totalSupply().intValue());

        var updated = planetNFTScore.tokenIdsOf(holder, 0, 100);
        var updatedTokenIds = updated.tokenIds;
        boolean burningSuccess = true;
        for (var tokenId : updatedTokenIds) {
            if (burned.equals(tokenId)) {
                burningSuccess = false;
                break;
            }
        }
        if (expected) {
            assertTrue(burningSuccess);
        } else {
            assertFalse(burningSuccess);
        }
    }

    @Test
    void burn() throws Exception {
        Address holder = KeyWallet.create().getAddress();
        _mintAndCheckBalance(planetScoreOwner, holder); // success test
        _mintAndCheckBalance(planetScoreOwner, holder); // success test
        // mint 2

        _burnAndCheckBalance(wallets[2], holder, false);
        _burnAndCheckBalance(planetScoreOwner, holder, true);
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

    private PlanetNFTScore.TokenInfo getFirstTokenInfo(Address address) throws Exception {
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
        return new PlanetNFTScore.TokenInfo(hPrice, isCompany, isPrivate, owner, uPrice, null);
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
            var tokenInfo = getFirstTokenInfo(wallet.getAddress());

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

    @Test
    void planetSupply() throws Exception {
        var totalSupply = planetNFTScore.totalSupply().intValue();
        int testMintCnt = 200;
        List<Bytes> txHashes = new ArrayList<>();
        Address newAddress = KeyWallet.create().getAddress();
        for (int i = 0; i < testMintCnt; i++) {
            Bytes txHash = planetNFTScore.mintPlanet(planetScoreOwner, newAddress, 1, BigInteger.ONE, BigInteger.ONE);
            txHashes.add(txHash);
        }

        for (int i = 0; i < testMintCnt; i++) {
            Bytes txHash = txHashes.get(i);
            assertSuccess(txHash);
        }
        var diff = planetNFTScore.totalSupply().intValue() - totalSupply;
        assertEquals(diff, testMintCnt);

        var tokenIds = planetNFTScore.tokenIdsOf(newAddress, 0, 2);
        for (var id : tokenIds.tokenIds) {
            var txHash = planetNFTScore.burn(planetScoreOwner, id);
            assertEquals(BigInteger.ONE, txHandler.getResult(txHash).getStatus());
            assertEquals(--testMintCnt, planetNFTScore.totalSupply().intValue() - totalSupply);
        }
        LOG.info("token supply(" + testMintCnt + ")");
        Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(Utils.getHeightNext(5));
    }

    void mintWithTokenId(Wallet holder, BigInteger mintId, BigInteger usdtPrice, BigInteger hvhPrice, boolean expected) throws Exception {
        Bytes txHash = planetNFTScore.mintPlanet(
                planetScoreOwner, holder.getAddress(), PlanetNFTScore.PLANET_PUBLIC, usdtPrice, hvhPrice, mintId);
        var result = txHandler.getResult(txHash);
        if (expected) {
            assertSuccess(result);
        } else {
            assertFailure(result);
            return;
        }

        var tokenIds = planetNFTScore.tokenIdsOf(holder.getAddress(), 0, 1);
        var tokenId = tokenIds.tokenIds.get(0);
        assertEquals(mintId, tokenId);
        var object = planetNFTScore.infoOf(tokenId);
        var tokenInfo = PlanetNFTScore.toTokenInfo(object);
        assertFalse(tokenInfo.isCompany());
        assertFalse(tokenInfo.isPrivate());
        assertEquals(holder.getAddress(), tokenInfo.getOwner());
        assertEquals(usdtPrice, tokenInfo.getUsdtPrice());
        assertEquals(hvhPrice, tokenInfo.getHavahPrice());
        assertEquals(result.getBlockHeight(), tokenInfo.getHeight());
    }

    @Test
    void duplicateTokenId() throws Exception {
        BigInteger testTokenId = BigInteger.valueOf(2000000001);
        Wallet holder = KeyWallet.create();
        BigInteger usdtPrice = BigInteger.TWO;
        BigInteger hvhPrice = BigInteger.TWO;
        // mint with testNonce - success
        mintWithTokenId(holder, testTokenId, usdtPrice, hvhPrice, true);

        holder = KeyWallet.create();
        usdtPrice = usdtPrice.add(BigInteger.ONE);
        hvhPrice = hvhPrice.add(BigInteger.ONE);

        for (int i = 0; i < 2; i++) {
            mintWithTokenId(holder, testTokenId, usdtPrice, hvhPrice, i == 1);
            if (i == 0) {
                var txHash = planetNFTScore.burn(planetScoreOwner, testTokenId);
                assertSuccess(txHandler.getResult(txHash));
            }
        }
    }

    @Test
    void checkAdmin() throws Exception {
        Address holder = KeyWallet.create().getAddress();

        // mint and burn with owner - success
        _mintAndCheckBalance(planetScoreOwner, holder);
        _burnAndCheckBalance(planetScoreOwner, holder, true);

        // setAdmin to admin
        Wallet admin = wallets[2];
        var txHash = planetNFTScore.setAdmin(planetScoreOwner, admin.getAddress());
        assertSuccess(txHandler.getResult(txHash));
        // mint and burn with owner - failure
        _mintAndCheckBalance(planetScoreOwner, holder, false);
        // mint and burn with admin - success
        _mintAndCheckBalance(admin, holder, true);
        _burnAndCheckBalance(planetScoreOwner, holder, false);
        _burnAndCheckBalance(admin, holder, true);

        // setAdmin to owner
        txHash = planetNFTScore.setAdmin(planetScoreOwner, planetScoreOwner.getAddress());
        assertFailure(txHandler.getResult(txHash));

        txHash = planetNFTScore.setAdmin(admin, planetScoreOwner.getAddress());
        assertSuccess(txHandler.getResult(txHash));
        // mint and burn with admin - failure
        // mint and burn with owner - success
        _mintAndCheckBalance(admin, holder, false);
        _mintAndCheckBalance(planetScoreOwner, holder, true);
        _burnAndCheckBalance(admin, holder, false);
        _burnAndCheckBalance(planetScoreOwner, holder, true);
    }

    @Test
    void changeMintApprover() throws Exception {
        Address holder = KeyWallet.create().getAddress();
        Wallet caller = wallets[1];

        // mint and burn with owner - success
        _mintAndCheckBalance(caller, holder, false);
        var txHash = planetNFTScore.addMintingApprover(planetScoreOwner, caller.getAddress());
        assertSuccess(txHandler.getResult(txHash));
        _mintAndCheckBalance(caller, holder, true);
        _mintAndCheckBalance(planetScoreOwner, holder, true);
        var approvers = planetNFTScore.getMintingApprover();
        assertEquals(approvers.size(), 1);
        assertEquals(approvers.get(0), caller.getAddress());

        Wallet caller2 = wallets[2];
        _mintAndCheckBalance(caller2, holder, false);
        txHash = planetNFTScore.addMintingApprover(planetScoreOwner, caller2.getAddress());
        assertSuccess(txHandler.getResult(txHash));
        _mintAndCheckBalance(caller2, holder, true);
        _mintAndCheckBalance(caller, holder, true);
        _mintAndCheckBalance(planetScoreOwner, holder, true);
        approvers = planetNFTScore.getMintingApprover();
        assertEquals(approvers.size(), 2);
        for (int i = 0; i < approvers.size(); i++) {
            if (i == 0) {
                assertEquals(approvers.get(0), caller.getAddress());
            } else {
                assertEquals(approvers.get(1), caller2.getAddress());
            }
        }

        txHash = planetNFTScore.removeMintingApprover(planetScoreOwner, caller.getAddress());
        assertSuccess(txHandler.getResult(txHash));
        approvers = planetNFTScore.getMintingApprover();
        assertEquals(1, approvers.size());
        assertEquals(approvers.get(0), caller2.getAddress());

        _mintAndCheckBalance(caller, holder, false);
        _mintAndCheckBalance(caller2, holder, true);
        _mintAndCheckBalance(planetScoreOwner, holder, true);


        txHash = planetNFTScore.removeMintingApprover(planetScoreOwner, caller2.getAddress());
        assertSuccess(txHandler.getResult(txHash));
        _mintAndCheckBalance(caller, holder, false);
        _mintAndCheckBalance(caller2, holder, false);
        _mintAndCheckBalance(planetScoreOwner, holder, true);
        approvers = planetNFTScore.getMintingApprover();
        assertEquals(approvers.size(), 0);
    }

//    @Test
//    void checkLimitTokenId() throws Exception {
//        Wallet holder = KeyWallet.create();
//        int max = Integer.MAX_VALUE;
//        var txHash = planetNFTScore.mintPlanet(holder.getAddress(), 2, BigInteger.ONE, BigInteger.ONE, BigInteger.valueOf(max));
//        assertSuccess(txHash);
//
//        txHash = planetNFTScore.mintPlanet(holder.getAddress(), 2, BigInteger.ONE, BigInteger.ONE, BigInteger.valueOf(max).add(BigInteger.ONE));
//        assertSuccess(txHash);
//
//        var index = planetNFTScore.tokenByIndex(BigInteger.valueOf(max).add(BigInteger.ONE));
//    }

    @Test
    void mintMultipleTokens() throws Exception {
        int tokenOffset = 5300;
        while (true) {
            boolean success = true;
            for (int i = 0; i < 5; i++) {
                var testTokenId = BigInteger.valueOf(tokenOffset + i);
                try {
                    planetNFTScore.ownerOf(testTokenId);
                } catch (Exception e) {
                    System.out.println("Excpetion (" + e + ")");
                    continue;
                }
                success = false;
            }
            if (success) {
                break;
            }
            tokenOffset += 1000;
        }

        Wallet holder = KeyWallet.create();
        var balance = planetNFTScore.balanceOf(holder.getAddress());
        assertEquals(0, balance.compareTo(BigInteger.ZERO));

        // mint multiple tokens in one tx
        var txHash = planetNFTScore.mintPlanet(planetScoreOwner,
                List.of(PlanetNFTScore.mintInfo(BigInteger.valueOf(tokenOffset++), BigInteger.ONE, BigInteger.ONE),
                        PlanetNFTScore.mintInfo(BigInteger.valueOf(tokenOffset++), BigInteger.ONE, BigInteger.ONE)),
                holder.getAddress(), PlanetNFTScore.PLANET_PRIVATE);
        assertSuccess(txHash);

        balance = planetNFTScore.balanceOf(holder.getAddress());
        assertEquals(2, balance.intValue());

        txHash = planetNFTScore.mintPlanet(planetScoreOwner,
                List.of(PlanetNFTScore.mintInfo(BigInteger.valueOf(tokenOffset), BigInteger.ONE, BigInteger.ONE),
                        PlanetNFTScore.mintInfo(BigInteger.valueOf(tokenOffset), BigInteger.ONE, BigInteger.ONE)),
                holder.getAddress(), PlanetNFTScore.PLANET_PRIVATE);
        assertFailure(txHash);
        balance = planetNFTScore.balanceOf(holder.getAddress());
        assertEquals(2, balance.intValue());

        txHash = planetNFTScore.mintPlanet(planetScoreOwner,
                List.of(PlanetNFTScore.mintInfo(BigInteger.valueOf(tokenOffset - 1), BigInteger.ONE, BigInteger.ONE)),
                holder.getAddress(), PlanetNFTScore.PLANET_PRIVATE);
        assertFailure(txHash);
        balance = planetNFTScore.balanceOf(holder.getAddress());
        assertEquals(2, balance.intValue());

        txHash = planetNFTScore.mintPlanet(planetScoreOwner,
                List.of(PlanetNFTScore.mintInfo(BigInteger.valueOf(tokenOffset - 1), BigInteger.ONE, BigInteger.ONE),
                        PlanetNFTScore.mintInfo(BigInteger.valueOf(tokenOffset), BigInteger.ONE, BigInteger.ONE)),
                holder.getAddress(), PlanetNFTScore.PLANET_PRIVATE);
        assertFailure(txHash);
        balance = planetNFTScore.balanceOf(holder.getAddress());
        assertEquals(2, balance.intValue());

        txHash = planetNFTScore.mintPlanet(planetScoreOwner,
                List.of(PlanetNFTScore.mintInfo(BigInteger.valueOf(tokenOffset), BigInteger.ONE, BigInteger.ONE)),
                holder.getAddress(), PlanetNFTScore.PLANET_PRIVATE);
        assertSuccess(txHash);
        balance = planetNFTScore.balanceOf(holder.getAddress());
        assertEquals(3, balance.intValue());
    }
}
