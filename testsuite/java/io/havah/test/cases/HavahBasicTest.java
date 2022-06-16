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
import java.util.List;

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
        planetNFTScore = new PlanetNFTScore(txHandler);

        try {
            Bytes txHash = txHandler.transfer(chain.godWallet, governorWallet.getAddress(), ICX);
            assertSuccess(txHandler.getResult(txHash));
        } catch (Exception ex) {
            fail(ex.getMessage());
        }
    }

    @AfterAll
    public static void clean() {

    }

    public BigInteger _getHeight() throws IOException {
        Block lastBlk = iconService.getLastBlock().execute();
        return lastBlk.getHeight();
    }

    public void _checkPlanetManager(Wallet wallet, Address address, boolean success) throws Exception {
        LOG.infoEntering("addPlanetManager");
        TransactionResult result = govScore.addPlanetManager(wallet, address);
        assertEquals(success ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");
        LOG.infoExiting();

        LOG.infoEntering("isPlanetManager");
        assertEquals(chainScore.isPlanetManager(address), success);
        LOG.infoExiting();
    }

    public void _mintPlanetNFT(Wallet wallet, Address to, int type, boolean success) throws Exception {
        Bytes txHash = planetNFTScore.mintPlanet(wallet, to, type, BigInteger.ONE, BigInteger.ONE);
        TransactionResult result = planetNFTScore.getResult(txHash);
        assertEquals(success ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");
    }

    public void _checkAndMintPlanetNFT(Address to) throws Exception {
        LOG.infoEntering("_checkAndMintPlanetNFT", "mint public PlanetNFT");
        var oldBalance = planetNFTScore.balanceOf(to).intValue();
        var oldTotalSupply = planetNFTScore.totalSupply().intValue();
        LOG.info("PlanetNFT Balance : " + oldBalance);
        LOG.info("PlanetNFT totalSupply : " + oldTotalSupply);

        _mintPlanetNFT(governorWallet, to, 2, true);

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

    public BigInteger _startRewardIssue(Wallet wallet, BigInteger addHeight, boolean success) throws IOException, ResultTimeoutException {
        LOG.infoEntering("_startRewardIssue", "expect : " + success);
        var height = _getHeight();
        var reward = height.add(addHeight);
        LOG.info("cur height : " + height);
        LOG.info("reward height : " + reward);
        TransactionResult result = govScore.startRewardIssue(wallet, reward);
        assertEquals(success ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");
        LOG.infoExiting();
        return reward;
    }

    public List<BigInteger> _tokenIdsOf(Address address, int idCount, BigInteger balance) throws Exception {
        LOG.infoEntering("_tokenIdsOf");
        PlanetNFTScore.TokenIds ids = planetNFTScore.tokenIdsOf(address, 0, idCount);
        assertEquals(idCount, ids.tokenIds.size());
        assertEquals(balance, ids.balance);
        LOG.infoExiting();
        return ids.tokenIds;
    }

    public void _reportPlanetWork(Wallet wallet, BigInteger planetId, boolean success) throws Exception {
        LOG.infoEntering("reportPlanetWork");
        TransactionResult result = chainScore.reportPlanetWork(wallet, planetId);
        assertEquals(success ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");
        LOG.infoExiting();
    }

    public void _checkAndClaimPlanetReward(Wallet wallet, BigInteger[] planetIds, boolean success) throws Exception {
        LOG.infoEntering("_checkAndClaimPlanetReward");
        LOG.info("planet balance (before claim) : " + txHandler.getBalance(wallet.getAddress()));
        _claimPlanetReward(wallet, planetIds, success);
        LOG.info("planet balance (after claim) : " + txHandler.getBalance(wallet.getAddress()));
        LOG.infoExiting();
    }

    public void _claimPlanetReward(Wallet wallet, BigInteger[] planetIds, boolean success) throws Exception {
        TransactionResult result = chainScore.claimPlanetReward(wallet, planetIds);
        assertEquals(success ? 1 : 0, result.getStatus().intValue(), "failure result(" + result + ")");
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
            LOG.info("claimable : " + obj.getItem("claimable").asInteger());
            LOG.info("height : " + obj.getItem("height").asInteger());
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        LOG.infoExiting();
    }

    @Test
    public void addPlanetManagerTest() throws Exception {
        LOG.infoEntering("getPlanetInfoTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet EOAWallet = KeyWallet.create();
        _checkPlanetManager(EOAWallet, planetManagerWallet.getAddress(), false);
        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        LOG.infoExiting();
    }

    @Test
    public void getPlanetInfoTest() throws Exception {
        LOG.infoEntering("getPlanetInfoTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress());
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);
        _getPlanetInfo(planetIds.get(0));
        _getPlanetInfo(BigInteger.valueOf(-1));
        LOG.infoExiting();
    }

    @Test
    public void getIssueInfoTest() throws Exception {
        LOG.infoEntering("getIssueInfoTest");
        _getIssueInfo();
        LOG.infoExiting();
    }

    @Test
    public void getRewardInfoTest() throws Exception {
        LOG.infoEntering("getRewardInfoTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress());
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);
        _getRewardInfo(planetIds.get(0));
        _getRewardInfo(BigInteger.valueOf(-1));
        LOG.infoExiting();
    }

    @Test
    public void setRewardIssueTest() throws Exception {
        LOG.infoEntering("setRewardIssueTest");
        KeyWallet EOAWallet = KeyWallet.create();
        _startRewardIssue(EOAWallet, BigInteger.valueOf(4), false);
        _startRewardIssue(governorWallet, BigInteger.valueOf(4), true);
        LOG.infoExiting();
    }

    @Test
    public void reportPlanetWorkTest() throws Exception {
        LOG.infoEntering("reportPlanetWorkTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        KeyWallet EOAWallet = KeyWallet.create();

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress());
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        _reportPlanetWork(EOAWallet, planetIds.get(0), false);
        _reportPlanetWork(planetManagerWallet, BigInteger.valueOf(-1), false);
        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        LOG.infoExiting();
    }

    public void _waitUtil(BigInteger height) throws Exception {
        var now = _getHeight();
        while (now.compareTo(height) < 0) {
            LOG.info("..");
            Thread.sleep(1500);
            now = _getHeight();
        }
    }
    @Test
    public void claimPlanetRewardTest() throws Exception {
        LOG.infoEntering("claimPlanetRewardTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();

        BigInteger rewardAdd = BigInteger.valueOf(4);

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress());
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);
        BigInteger waitHeight = _startRewardIssue(governorWallet, rewardAdd, true);
        _getRewardInfo(planetIds.get(0));
        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);

        _waitUtil(waitHeight);

        _getRewardInfo(planetIds.get(0));
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true);

        // mint second planet nft
        _checkAndMintPlanetNFT(planetWallet.getAddress());
        planetIds = _tokenIdsOf(planetWallet.getAddress(), 2, BigInteger.TWO);
        _getRewardInfo(planetIds.get(1));

        _waitUtil(_getHeight().add(rewardAdd));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        _reportPlanetWork(planetManagerWallet, planetIds.get(1), true);

        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0), planetIds.get(1)}, true);

        LOG.infoExiting();
    }
}
