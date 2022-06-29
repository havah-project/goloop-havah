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
import org.junit.jupiter.api.*;

import java.io.IOException;
import java.math.BigInteger;
import java.util.List;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.fail;

@Tag(Constants.TAG_HAVAH)
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
public class HavahBasicTest extends TestBase {
    private static TransactionHandler txHandler;
    private static IconService iconService;
    private static GovScore govScore;
    private static ChainScore chainScore;
    private static KeyWallet governorWallet;
    private static PlanetNFTScore planetNFTScore;

    private static final int PLANETTYPE_PUBLIC = 2;
    private static final int PLANETTYPE_PRIVATE = 1;
    private static final int PLANETTYPE_COMPANY = 4;

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
        planetNFTScore = new PlanetNFTScore(governorWallet, txHandler);

        try {
            Bytes txHash = txHandler.transfer(chain.godWallet, governorWallet.getAddress(), ICX);
            assertSuccess(txHandler.getResult(txHash));
        } catch (Exception ex) {
            fail(ex.getMessage());
        }

        _waitUtil(_startRewardIssue(governorWallet, BigInteger.valueOf(5), true));
    }

    @AfterAll
    public static void clean() {

    }

    private static BigInteger _getHeight() throws IOException {
        Block lastBlk = iconService.getLastBlock().execute();
        return lastBlk.getHeight();
    }

    private static BigInteger _getStartHeightOfTerm(BigInteger term, BigInteger termPeriod, BigInteger rewardStartHeight) {
        return term.subtract(BigInteger.ONE).multiply(termPeriod).add(rewardStartHeight);
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

    public void _checkAndMintPlanetNFT(Address to, int type) throws Exception {
        LOG.infoEntering("_checkAndMintPlanetNFT", "mint PlanetNFT type : " + type);
        var oldBalance = planetNFTScore.balanceOf(to).intValue();
        var oldTotalSupply = planetNFTScore.totalSupply().intValue();
        LOG.info("PlanetNFT Balance : " + oldBalance);
        LOG.info("PlanetNFT totalSupply : " + oldTotalSupply);

        _mintPlanetNFT(governorWallet, to, type, true);

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

    private static BigInteger _startRewardIssue(Wallet wallet, BigInteger addHeight, boolean success) throws IOException, ResultTimeoutException {
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

    public BigInteger _getPlanetInfo(BigInteger planetId) throws IOException {
        BigInteger regHeight = BigInteger.ZERO;
        LOG.infoEntering("_getPlanetInfo", "planetId : " + planetId);
        try {
            RpcObject obj = chainScore.getPlanetInfo(planetId);
            LOG.info("owner : " + obj.getItem("owner").asAddress());
            LOG.info("usdtPrice : " + obj.getItem("usdtPrice").asInteger());
            LOG.info("havahPrice : " + obj.getItem("havahPrice").asInteger());
            LOG.info("isCompany : " + obj.getItem("isCompany").asBoolean());
            LOG.info("isPrivate : " + obj.getItem("isPrivate").asBoolean());
            LOG.info("height : " + obj.getItem("height").asInteger());
            regHeight = obj.getItem("height").asInteger();
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        LOG.infoExiting();
        return regHeight;
    }

    public BigInteger _getIssueInfo() throws IOException {
        BigInteger issueStart = BigInteger.ZERO;
        LOG.infoEntering("_getPlanetInfo");
        RpcObject obj = chainScore.getIssueInfo();
        LOG.info("termPeriod : " + obj.getItem("termPeriod").asInteger());
        LOG.info("issueReductionCycle : " + obj.getItem("issueReductionCycle").asInteger());
        LOG.info("height : " + obj.getItem("height").asInteger());
        try {
            LOG.info("termSequence : " + obj.getItem("termSequence").asInteger());
            issueStart = obj.getItem("issueStart").asInteger();
            LOG.info("issueStart : " + issueStart);
        } catch (NullPointerException e) {
            LOG.info("startRewardIssue not called. termSequence and issueStart is null.");
        }
        LOG.infoExiting();
        return issueStart;
    }

    public BigInteger _getRewardInfo(BigInteger planetId) throws Exception {
        BigInteger claimable = BigInteger.ZERO;
        LOG.infoEntering("_getRewardInfo", "planetId : " + planetId);
        try {
            RpcObject obj = chainScore.getRewardInfo(planetId);
            LOG.info("total : " + obj.getItem("total").asInteger());
            LOG.info("remain : " + obj.getItem("remain").asInteger());
            LOG.info("claimable : " + obj.getItem("claimable").asInteger());
            LOG.info("height : " + obj.getItem("height").asInteger());
            claimable = obj.getItem("claimable").asInteger();
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        LOG.infoExiting();
        return claimable;
    }

    public BigInteger _getIssueReductionCycle() throws Exception {
        RpcObject obj = chainScore.getIssueInfo();
        return obj.getItem("issueReductionCycle").asInteger();
    }

    private static void _waitUtil(BigInteger height) throws Exception {
        var now = _getHeight();
        while (now.compareTo(height) < 0) {
            LOG.info("now(" + now + ") wait(" + height + ")");
            Thread.sleep(1500);
            now = _getHeight();
        }
    }

    private static void _waitUtilNextHeight() throws Exception {
        var now = _getHeight();
        var termPeriod = _getTermPeriod();
        var height = now.add(termPeriod.subtract(now.mod(termPeriod)));
        while (now.compareTo(height) < 0) {
            LOG.info("now(" + now + ") wait(" + height + ")");
            Thread.sleep(1500);
            now = _getHeight();
        }
    }

    public static BigInteger _getTermPeriod() throws IOException {
        // termPeriod : 주기당 블록 수 (Blocks) 기본 : 43200 (하루)
        RpcObject obj = chainScore.getIssueInfo();
        return obj.getItem("termPeriod").asInteger();
    }

    public BigInteger _getCurrentPublicReward() throws IOException {
        //BigInteger reward = Constants.INITIAL_ISSUE_AMOUNT.divide(BigInteger.valueOf(100));
        BigInteger reward = Constants.INITIAL_ISSUE_AMOUNT;
        try {
            RpcObject obj = chainScore.getIssueInfo();
            BigInteger termSequence = obj.getItem("termSequence").asInteger();
            BigInteger issueReductionCycle = obj.getItem("issueReductionCycle").asInteger();

            int count = termSequence.divide(issueReductionCycle).intValue();
            for(int i=0; i<count; i++) {
                reward = reward.multiply(BigInteger.valueOf(7)).divide(BigInteger.TEN);
            }
        } catch (NullPointerException e) {
            LOG.info("startRewardIssue not called. termSequence and issueStart is null.");
            return BigInteger.ZERO;
        }

        return reward.divide(planetNFTScore.totalSupply());
    }


    @Test
    @Order(1)
    public void addPlanetManagerTest() throws Exception {
        LOG.infoEntering("addPlanetManagerTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet EOAWallet = KeyWallet.create();
        _checkPlanetManager(EOAWallet, planetManagerWallet.getAddress(), false);
        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        LOG.infoExiting();
    }

    @Test
    @Order(2)
    public void getPlanetInfoTest() throws Exception {
        LOG.infoEntering("getPlanetInfoTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PUBLIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);
        _getPlanetInfo(planetIds.get(0));
        _getPlanetInfo(BigInteger.valueOf(-1));
        LOG.infoExiting();
    }

    @Test
    @Order(3)
    public void getIssueInfoTest() throws Exception {
        LOG.infoEntering("getIssueInfoTest");
        _getIssueInfo();
        LOG.infoExiting();
    }

    @Test
    @Order(4)
    public void getRewardInfoTest() throws Exception {
        LOG.infoEntering("getRewardInfoTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PUBLIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);
        _getRewardInfo(planetIds.get(0));
        _getRewardInfo(BigInteger.valueOf(-1));
        LOG.infoExiting();
    }

//    @Test
//    @Order(5)
//    public void setRewardIssueTest() throws Exception {
//        LOG.infoEntering("setRewardIssueTest");
//        KeyWallet EOAWallet = KeyWallet.create();
//        BigInteger startReward = BigInteger.valueOf(5);
//        _startRewardIssue(EOAWallet, startReward, false);
//        BigInteger reward = _startRewardIssue(governorWallet, startReward, true);
//        if(_getHeight().compareTo(reward) < 0)
//            _startRewardIssue(governorWallet, startReward, true); // 리워드가 시작되기 전 호출은 성공해야함.
//        else
//            LOG.info("reward already started. ignore continuous call test.");
//
//        _waitUtil(reward);
//
//        _startRewardIssue(governorWallet, startReward, false); // 리워드가 시작되면 실패해야 한다고 함.
//        LOG.infoExiting();
//    }

    @Test
    @Order(6)
    public void reportPlanetWorkTest() throws Exception {
        LOG.infoEntering("reportPlanetWorkTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        KeyWallet EOAWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PUBLIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        _reportPlanetWork(EOAWallet, planetIds.get(0), false);
        _reportPlanetWork(planetManagerWallet, BigInteger.valueOf(-1), false);

        _waitUtil(_getHeight().add(termPeriod));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        LOG.infoExiting();
    }

    @Test
    @Order(7)
    public void claimPublicPlanetRewardTest() throws Exception {
        LOG.infoEntering("claimPublicPlanetRewardTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PUBLIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        _waitUtil(_getHeight().add(termPeriod));
        _getPlanetInfo(planetIds.get(0));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        BigInteger claimable = _getRewardInfo(planetIds.get(0));
        BigInteger expected = _getCurrentPublicReward();
        LOG.info("expected = " + expected);
        assertEquals(0, claimable.compareTo(expected), "claimable is not expected");
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true);
        _getRewardInfo(planetIds.get(0));

        // mint second planet nft
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PUBLIC);
        planetIds = _tokenIdsOf(planetWallet.getAddress(), 2, BigInteger.TWO);

        _waitUtil(_getHeight().add(termPeriod));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        _reportPlanetWork(planetManagerWallet, planetIds.get(1), true);

        claimable = _getRewardInfo(planetIds.get(0));
        expected = _getCurrentPublicReward();
        LOG.info("expected = " + expected);
        assertEquals(0, claimable.compareTo(expected), "claimable is not expected");

        claimable = _getRewardInfo(planetIds.get(1));
        LOG.info("expected = " + expected);
        assertEquals(0, claimable.compareTo(expected), "claimable is not expected");

        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0), planetIds.get(1)}, true);

        LOG.infoExiting();
    }

    @Test
    @Order(9)
    public void claimCompanyPlanetRewardTest() throws Exception {
        LOG.infoEntering("claimCompanyPlanetRewardTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();
        LOG.info("termPeriod : " + termPeriod);

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_COMPANY);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        _getPlanetInfo(planetIds.get(0));
        LOG.info("planetWallet : " + planetWallet.getAddress());

        _waitUtil(_getHeight().add(termPeriod));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        BigInteger claimable = _getRewardInfo(planetIds.get(0));
        BigInteger reward = _getCurrentPublicReward();
        BigInteger expectedPlanet = reward.multiply(BigInteger.valueOf(4)).divide(BigInteger.TEN);
        BigInteger expectedEco = reward.multiply(BigInteger.valueOf(6)).divide(BigInteger.TEN);
        LOG.info("reward = " + reward);
        LOG.info("Planet = " + expectedPlanet);
        LOG.info("Eco = " + expectedEco);

        assertEquals(0, claimable.compareTo(expectedPlanet), "claimable is not expected");

        BigInteger beforeEco = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
        LOG.info("ecosystem balance (before claim) : " + beforeEco);
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true);

        _waitUtilNextHeight();

        BigInteger afterEco = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
        LOG.info("ecosystem balance (after claim) : " + afterEco);
        assertEquals(0, afterEco.subtract(beforeEco).compareTo(expectedEco), "ecosystem reward is not expected");

        LOG.infoExiting();
    }

    @Test
    @Order(10)
    public void issueReductionCycleTest() throws Exception {
        LOG.infoEntering("issueReductionCycleTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();
        BigInteger issueReductionCycle = _getIssueReductionCycle();

        LOG.info("termPeriod : " + termPeriod);
        LOG.info("issueReductionCycle : " + issueReductionCycle);

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PUBLIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        _getPlanetInfo(planetIds.get(0));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        BigInteger claimable = _getRewardInfo(planetIds.get(0));
        BigInteger termReward = _getCurrentPublicReward();
        LOG.info("termReward : " + termReward);
        assertEquals(claimable.compareTo(termReward), 0, "term reward is not equals to claimable");
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true);

        _waitUtil(_getIssueInfo().add(termPeriod.multiply(issueReductionCycle)));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        claimable = _getRewardInfo(planetIds.get(0));
        termReward = _getCurrentPublicReward();
        LOG.info("reduction termReward : " + termReward);
        assertEquals(claimable.compareTo(termReward), 0, "reduction term reward is not equals to claimable");
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true);

        LOG.infoExiting();
    }
}