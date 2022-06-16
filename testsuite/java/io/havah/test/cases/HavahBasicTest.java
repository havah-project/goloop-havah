package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Block;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.http.HttpProvider;
import foundation.icon.icx.transport.jsonrpc.RpcError;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.test.common.Env;
import foundation.icon.test.common.ResultTimeoutException;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.score.ChainScore;
import io.havah.test.score.GovScore;
import io.havah.test.score.PlanetNFTScore;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.fail;

@Tag(Constants.TAG_HAVAH)
public class HavahBasicTest extends TestBase {
    private static TransactionHandler txHandler;
    private static IconService iconService;
    private static GovScore govScore;
    private static ChainScore chainScore;
    private static KeyWallet governorWallet;
    private static KeyWallet planetManagerWallet;
    private static KeyWallet planetWallet;
    private static KeyWallet EOAWallet;
    private static PlanetNFTScore planetNFTScore;

    @BeforeAll
    public static void setup() throws Exception {
        Env.Node node = Env.nodes[0];
        Env.Channel channel = node.channels[0];
        Env.Chain chain = channel.chain;
        iconService = new IconService(new HttpProvider(channel.getAPIUrl(Env.testApiVer)));
        txHandler = new TransactionHandler(iconService, chain);

        govScore = new GovScore(txHandler);
        governorWallet = chain.governorWallet;

        chainScore = new ChainScore(txHandler);

        try {
            Bytes txHash = txHandler.transfer(chain.godWallet, governorWallet.getAddress(), ICX);
            assertSuccess(txHandler.getResult(txHash));

            planetManagerWallet = KeyWallet.create();
            planetWallet = KeyWallet.create();
            EOAWallet = KeyWallet.create();
        } catch (Exception ex) {
            fail(ex.getMessage());
        }

        planetNFTScore = new PlanetNFTScore(txHandler);
    }

    @AfterAll
    public static void clean() {

    }

    public BigInteger _getHeight() throws IOException {
        Block lastBlk = iconService.getLastBlock().execute();
        return lastBlk.getHeight();
    }

    public void _checkPlanetManager(Wallet wallet) throws Exception {
        LOG.infoEntering("addPlanetManager");
        Address address = wallet.getAddress();
        assertFailure(govScore.addPlanetManager(EOAWallet, address));
        assertSuccess(govScore.addPlanetManager(governorWallet, address));
        LOG.infoExiting();

        LOG.infoEntering("isPlanetManager");
        assertEquals(chainScore.isPlanetManager(address), true);
        LOG.infoExiting();
    }

    public void _mintPlanetNFT(Wallet wallet, Address to, int type, boolean success) throws Exception {
        Bytes txHash = planetNFTScore.mintPlanet(wallet, to, type, BigInteger.ONE, BigInteger.ONE);
        TransactionResult result = planetNFTScore.getResult(txHash);
        assertEquals(success ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");
    }

    public void _checkAndMintPlanetNFT(Address to) throws Exception {
        LOG.infoEntering("mintPlanet", "mint public PlanetNFT");
        var oldBalance = planetNFTScore.balanceOf(to).intValue();
        var oldTotalSupply = planetNFTScore.totalSupply().intValue();
        LOG.info("PlanetNFT Balance : " + oldBalance);
        LOG.info("PlanetNFT totalSupply : " + oldTotalSupply);

        _mintPlanetNFT(governorWallet, to, 1, true);

        // compare nft balance
        var balance = planetNFTScore.balanceOf(to).intValue();
        assertEquals(oldBalance + 1, balance);
        LOG.info("PlanetNFT Balance : " + balance);

        // compare nft supply
        var totalSupply = planetNFTScore.totalSupply().intValue();
        assertEquals(oldTotalSupply + 1, totalSupply);
        LOG.info("PlanetNFT totalSupply : " + totalSupply);
        LOG.infoExiting();
    }

    public BigInteger _startRewardIssue() throws IOException, ResultTimeoutException {
        LOG.infoEntering("startRewardIssue");
        var height = _getHeight();
        var reward = height.add(BigInteger.valueOf(4));
        LOG.info("cur height : " + height);
        LOG.info("reward height : " + reward);
        assertFailure(govScore.startRewardIssue(EOAWallet, reward));
        assertSuccess(govScore.startRewardIssue(governorWallet, reward));
        LOG.infoExiting();
        return reward;
    }

    public BigInteger _tokenIdsOf(Address address) throws Exception {
        LOG.infoEntering("_tokenIdsOf");
        PlanetNFTScore.TokenIds ids = planetNFTScore.tokenIdsOf(address, 0, 1);
        assertEquals(1, ids.tokenIds.size());
        assertEquals(BigInteger.ONE, ids.balance);
        LOG.infoExiting();
        return ids.tokenIds.get(0);
    }

    public void _reportPlanetWork(Wallet wallet, BigInteger planetId, boolean success) throws Exception {
        LOG.infoEntering("reportPlanetWork");
        TransactionResult result = chainScore.reportPlanetWork(wallet, planetId);
        assertEquals(success ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");
        LOG.infoExiting();
    }

    public void _checkAndClaimPlanetReward(Wallet wallet, BigInteger planetId, boolean success) throws Exception {
        LOG.infoEntering("_checkAndClaimPlanetReward");
        LOG.info("planet balance (before claim) : " + txHandler.getBalance(wallet.getAddress()).intValue());
        _claimPlanetReward(wallet, planetId, success);
        LOG.info("planet balance (after claim) : " + txHandler.getBalance(wallet.getAddress()).intValue());
        LOG.infoExiting();
    }

    public void _claimPlanetReward(Wallet wallet, BigInteger planetId, boolean success) throws Exception {
        try {
            TransactionResult result = chainScore.claimPlanetReward(wallet, new BigInteger[]{planetId});
            assertEquals(success ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_PENDING, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
    }

    public void _getPlanetInfo(BigInteger planetId) throws IOException {
        LOG.infoEntering("_getPlanetInfo", "planetId : " + planetId);
        try {
            RpcObject obj = chainScore.getPlanetInfo(planetId);
            LOG.info("owner : " + obj.getItem("owner").asAddress());
            LOG.info("usdtPrice : " + obj.getItem("usdtPrice").asInteger());
            LOG.info("havahPrice : " + obj.getItem("havahPrice").asInteger());
            LOG.info("isCompany : " + obj.getItem("isCompany").asBoolean());
            LOG.info("isPrivate : " + obj.getItem("isPrivate").asBoolean());
            LOG.info("height : " + obj.getItem("height").asInteger());
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        LOG.infoExiting();
    }

    public void _getIssueInfo() throws IOException {
        LOG.infoEntering("_getPlanetInfo");
        RpcObject obj = chainScore.getIssueInfo();
        LOG.info("termPeriod : " + obj.getItem("termPeriod").asInteger());
        LOG.info("issueReductionCycle : " + obj.getItem("issueReductionCycle").asInteger());
        LOG.info("height : " + obj.getItem("height").asInteger());
        try {
            LOG.info("termSequence : " + obj.getItem("termSequence").asInteger());
            LOG.info("issueStart : " + obj.getItem("issueStart").asInteger());
        } catch (NullPointerException e) {
            LOG.info("startRewardIssue not called. termSequence and issueStart is null.");
        }
        LOG.infoExiting();
    }

    public void _getRewardInfo(BigInteger planetId) throws Exception {
        LOG.infoEntering("_getRewardInfo", "planetId : " + planetId);
        try {
            RpcObject obj = chainScore.getRewardInfo(planetId);
            LOG.info("total : " + obj.getItem("total").asInteger());
            LOG.info("remain : " + obj.getItem("remain").asInteger());
            LOG.info("claimable : " + obj.getItem("height").asInteger());
            LOG.info("height : " + obj.getItem("height").asInteger());
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        LOG.infoExiting();
    }

//    @Test
    public void getPlanetInfoTest() throws Exception {
        LOG.infoEntering("getPlanetInfoTest");
        _checkPlanetManager(planetManagerWallet);
        _checkAndMintPlanetNFT(planetWallet.getAddress());
        BigInteger planetId = _tokenIdsOf(planetWallet.getAddress());
        _getPlanetInfo(planetId);
        _getPlanetInfo(BigInteger.valueOf(-1));
        LOG.infoExiting();
    }

//    @Test
    public void getIssueInfoTest() throws Exception {
        LOG.infoEntering("getIssueInfoTest");
        _getIssueInfo();
        LOG.infoExiting();
    }

//    @Test
    public void getRewardInfoTest() throws Exception {
        LOG.infoEntering("getRewardInfoTest");
        _checkPlanetManager(planetManagerWallet);
        _checkAndMintPlanetNFT(planetWallet.getAddress());
        BigInteger planetId = _tokenIdsOf(planetWallet.getAddress());
        _getRewardInfo(planetId);
        _getRewardInfo(BigInteger.valueOf(-1));
        LOG.infoExiting();
    }

    @Test
    public void test() throws Exception {
        LOG.infoEntering("test");
        _checkPlanetManager(planetManagerWallet);
        _checkAndMintPlanetNFT(planetWallet.getAddress());
        BigInteger planetId = _tokenIdsOf(planetWallet.getAddress());
        _startRewardIssue();
        _getRewardInfo(planetId);
//        _reportPlanetWork(governorWallet, planetId, false);
        _reportPlanetWork(planetManagerWallet, planetId, true);
//        LOG.info("cur Height" + _getHeight());
        LOG.info("waiting 11 sec....");
        Thread.sleep(11000);
//        LOG.info("cur Height" + _getHeight());
        _getRewardInfo(planetId);
        _reportPlanetWork(planetManagerWallet, planetId, true);
////        _reportPlanetWork(planetManagerWallet, planetId, false);
        _getRewardInfo(planetId);
        _checkAndClaimPlanetReward(planetWallet, planetId, true);
        LOG.infoExiting();
    }
}
