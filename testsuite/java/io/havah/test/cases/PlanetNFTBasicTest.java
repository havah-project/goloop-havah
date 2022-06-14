package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.http.HttpProvider;
import foundation.icon.icx.transport.jsonrpc.RpcItem;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.Env;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.score.PlanetNFTScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotEquals;


@Tag(Constants.TAG_HAVAH)
public class PlanetNFTBasicTest extends TestBase {
    private static KeyWallet[] wallets;
    private static KeyWallet deployer;
    private static PlanetNFTScore planetNFTScore;
    private static TransactionHandler txHandler;

    @BeforeAll
    static void setup() throws Exception {
        Env.Node node = Env.nodes[0];
        Env.Channel channel = node.channels[0];
        Env.Chain chain = channel.chain;
        IconService iconService = new IconService(new HttpProvider(channel.getAPIUrl(Env.testApiVer)));
        txHandler = new TransactionHandler(iconService, chain);
        deployer = txHandler.getChain().governorWallet;

        // init wallets
        wallets = new KeyWallet[]{
                KeyWallet.create(), KeyWallet.create(), KeyWallet.create()
        };
        List<KeyWallet> walletList = new ArrayList<>(Arrays.asList(wallets));
        walletList.add(deployer);
        BigInteger amount = ICX.multiply(BigInteger.valueOf(50));
        for (KeyWallet wallet : walletList) {
            txHandler.transfer(wallet.getAddress(), amount);
        }
//        for (KeyWallet wallet : walletList) {
//            ensureIcxBalance(txHandler, wallet.getAddress(), BigInteger.ZERO, amount);
//        }
        planetNFTScore = new PlanetNFTScore(txHandler);
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
        boolean success = wallet.equals(deployer);
        LOG.infoEntering("Mint", success ? "success" : "failure");

        // check balance of wallets[0] then compare after mint
        var balance = planetNFTScore.balanceOf(to).intValue();
        var totalSupply = planetNFTScore.totalSupply().intValue();
        Bytes txHash = planetNFTScore.mintPlanet(wallet, to, 1, BigInteger.ONE, BigInteger.ONE);
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

    @Test
    void mintPlanet() throws Exception {
        _mintAndCheckBalance(wallets[0]); // failure test
        _mintAndCheckBalance(deployer); // success test
    }

    @Test
    void mintPlanetAndCheckBalances() throws Exception {
        Address newAddress = KeyWallet.create().getAddress();
        assertEquals(0, planetNFTScore.balanceOf(newAddress).intValue());

        var totalSupply = planetNFTScore.totalSupply().intValue();
        int mintCnt = 5;
        for (int i = 0; i < mintCnt; i++) {
            _mintAndCheckBalance(deployer, newAddress); // negative test
        }

        assertEquals(mintCnt, planetNFTScore.balanceOf(newAddress).intValue());
        assertEquals(totalSupply + mintCnt, planetNFTScore.totalSupply().intValue());
        totalSupply += mintCnt;

        // TODO test with change type
        int testMintCnt = 1000;
        int failureCnt = 0;
        List<Bytes> txHashes = new ArrayList<>();
        for (int i = 0; i < testMintCnt; i++) {
            Bytes txHash = planetNFTScore.mintPlanet(deployer, newAddress, 1, BigInteger.ONE, BigInteger.ONE);
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
    }

    // mint & transfer
    void _mintAndTransfer(Wallet holder, Wallet from, Address to) throws Exception {
        boolean success = holder.equals(from);
        LOG.infoEntering("Transfer", success ? "success" : "failure");
        _mintAndCheckBalance(deployer, holder.getAddress());

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

    @Test
    void transfer() throws Exception {
        _mintAndTransfer(wallets[0], wallets[1], wallets[2].getAddress()); // failure case
        _mintAndTransfer(wallets[0], wallets[2], wallets[0].getAddress()); // failure case
        // check 내가 나에게 주는것....???
        _mintAndTransfer(wallets[0], wallets[0], wallets[1].getAddress()); // success case
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
        _mintAndCheckBalance(deployer, wallet[0].getAddress()); // mint with cnt
        _mintAndTransferFrom(wallet[0], wallet[1], wallet[2], wallet[3], wallet[4].getAddress()); // failure
        _mintAndTransferFrom(wallet[0], wallet[0], wallet[1], wallet[2], wallet[4].getAddress()); // failure
        _mintAndTransferFrom(wallet[0], wallet[0], wallet[1], wallet[1], wallet[4].getAddress()); // success
    }

    @Test
    void burn() throws Exception {
        Address holder = KeyWallet.create().getAddress();
        _mintAndCheckBalance(deployer, holder); // success test
        _mintAndCheckBalance(deployer, holder); // success test
        // mint 2
        int balance = planetNFTScore.balanceOf(holder).intValue();
        int supply = planetNFTScore.totalSupply().intValue();
        var tokenIdsMap = planetNFTScore.tokenIdsOf(holder, 0, 100);
        var tokenIds = tokenIdsMap.tokenIds;
        var burned = tokenIds.get(0);
        var txHash = planetNFTScore.burn(deployer, burned);
        var result = planetNFTScore.getResult(txHash);
        assertEquals(1, result.getStatus().intValue());
        assertEquals(balance - 1, planetNFTScore.balanceOf(holder).intValue());
        assertEquals(supply - 1, planetNFTScore.totalSupply().intValue());

        var updated = planetNFTScore.tokenIdsOf(holder, 0, 100);
        var updatedTokenIds = updated.tokenIds;
        for (var tokenId : updatedTokenIds) {
            assertNotEquals(burned.intValue(), tokenId.intValue());
        }
    }

    @Test
    void suppply() {

    }

//    @Test
//    void tokenInfo() throws IOException {
//        RpcObject params = new RpcObject.Builder()
//                .put("_tokenId", new RpcValue(BigInteger.ONE))
//                .build();
//        RpcItem item = planetNFTScore.call("infoOf", params);
//    }

    int _getOpState(BigInteger tokenId) throws Exception {
        RpcObject params = new RpcObject.Builder()
                .put("_tokenId", new RpcValue(tokenId))
                .build();
        RpcItem item = planetNFTScore.call("getOpState", params);
        return item.asInteger().intValue();
    }

    private static final int _agentNone = 0;
    private static final int _agentRequestingDelegation = 1;
    private static final int _agentAccepted = 2;
    private static final int _agentRequestingCancel = 3;

    void _assertResult(Bytes txHash, boolean success) throws Exception {
        TransactionResult result = planetNFTScore.getResult(txHash);
        // print result
        assertEquals(success ? 1 : 0, result.getStatus().intValue());
    }

    void _validStepForAgentRequest(Wallet requester, BigInteger tokenId, boolean successTest) throws Exception {
        String testCase = (successTest ? "success" : "failure") + " case";
        LOG.infoEntering("validStepForAgentRequest tokenId(" + tokenId + ") for " + testCase);
        boolean validRequester = requester.getAddress().equals(planetNFTScore.ownerOf(tokenId));
        assertEquals(successTest, validRequester);
        Wallet agent = KeyWallet.create();
        assertEquals(_agentNone, _getOpState(tokenId));
        TransactionResult result = planetNFTScore.getResult(planetNFTScore.requestStartOp(requester, tokenId, agent.getAddress()));
        assertEquals(validRequester ? 1 : 0, result.getStatus().intValue());
        assertEquals(validRequester ? _agentRequestingDelegation : _agentNone, _getOpState(tokenId));

        _assertResult(planetNFTScore.cancelStartOpReq(requester, tokenId), validRequester);
        assertEquals(_agentNone, _getOpState(tokenId));

        _assertResult(planetNFTScore.requestStartOp(requester, tokenId, agent.getAddress()), validRequester);
        assertEquals(validRequester ? _agentRequestingDelegation : _agentNone, _getOpState(tokenId));
        _assertResult(planetNFTScore.rejectStartOpReq(agent, tokenId), validRequester);
        assertEquals(validRequester ? 1 : 0, result.getStatus().intValue());
        assertEquals(_agentNone, _getOpState(tokenId));

        _assertResult(planetNFTScore.requestStartOp(requester, tokenId, agent.getAddress()), validRequester);
        assertEquals(validRequester ? _agentRequestingDelegation : _agentNone, _getOpState(tokenId));

        _assertResult(planetNFTScore.acceptStartOpReq(agent, tokenId, agent.getAddress()), validRequester);
        assertEquals(validRequester ? _agentAccepted : _agentNone, _getOpState(tokenId));

        // check planetInfo
        _assertResult(planetNFTScore.requestStopOp(requester, tokenId), validRequester);
        assertEquals(validRequester ? _agentRequestingCancel : _agentNone, _getOpState(tokenId));

        _assertResult(planetNFTScore.cancelStopOpReq(requester, tokenId), validRequester);
        assertEquals(validRequester ? _agentAccepted : _agentNone, _getOpState(tokenId));

        _assertResult(planetNFTScore.requestStopOp(requester, tokenId), validRequester);
        assertEquals(validRequester ? _agentRequestingCancel : _agentNone, _getOpState(tokenId));

        _assertResult(planetNFTScore.acceptStopOpReq(agent, tokenId), validRequester);
        assertEquals(_agentNone, _getOpState(tokenId));

        // check planetInfo
        LOG.infoExiting();
    }

    @Test
    void requestAgent() throws Exception {
        Wallet holder = wallets[0];
        _mintAndCheckBalance(deployer, holder.getAddress());
        var tokenIdsMap = planetNFTScore.tokenIdsOf(holder.getAddress(), 0, 10);
        var tokenIds = tokenIdsMap.tokenIds;
        var tokenId = tokenIds.get(0);
        Wallet fake = wallets[1];
        _validStepForAgentRequest(fake, tokenId, false);
        _validStepForAgentRequest(holder, tokenId, true);
    }

    // test for agent
    // reqeustStartOp
    //
    // approved
    // balanceOf
    // ownerOf
    // mint
    // burn
    // approve
    // transfer
    // transferFrom

    // test for agent state
}
