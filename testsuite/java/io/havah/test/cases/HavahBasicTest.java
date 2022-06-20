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

    private static final int PLANETTYPE_PULBIC = 2;
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

    public int _getRewardInfo(BigInteger planetId) throws Exception {
        int claimable = 0;
        LOG.infoEntering("_getRewardInfo", "planetId : " + planetId);
        try {
            RpcObject obj = chainScore.getRewardInfo(planetId);
            LOG.info("total : " + obj.getItem("total").asInteger());
            LOG.info("remain : " + obj.getItem("remain").asInteger());
            LOG.info("claimable : " + obj.getItem("claimable").asInteger());
            LOG.info("height : " + obj.getItem("height").asInteger());
            claimable = obj.getItem("claimable").asInteger().intValue();
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        LOG.infoExiting();
        return claimable;
    }

    public int _getClaimable(BigInteger planetId) throws Exception {
        int ret = 0;
        try {
            RpcObject obj = chainScore.getRewardInfo(planetId);
            ret = obj.getItem("claimable").asInteger().intValue();
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        return ret;
    }

    public void _waitUtil(BigInteger height) throws Exception {
        var now = _getHeight();
        while (now.compareTo(height) < 0) {
            LOG.info("..");
            Thread.sleep(1500);
            now = _getHeight();
        }
    }

    public BigInteger _getTermPeriod() throws IOException {
        // termPeriod : 주기당 블록 수 (Blocks) 기본 : 43200 (하루)
        RpcObject obj = chainScore.getIssueInfo();
        return obj.getItem("termPeriod").asInteger();
    }

    public BigInteger _getPrivateLockup() throws IOException {
        // privateLockup : 초기판매 보상 잠금기간 (텀단위) 기본 : 360 (1년)
        RpcObject obj = chainScore.getIssueInfo();
        return obj.getItem("privateLockup").asInteger();
    }

    public BigInteger _getPrivateReleaseCycle() throws IOException {
        // privateReleaseCycle : 초기판매 보상 분배주기 (텀단위) 기본 : 30 (1달)
        RpcObject obj = chainScore.getIssueInfo();
        return obj.getItem("privateReleaseCycle").asInteger();
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
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PULBIC);
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
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PULBIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);
        _getRewardInfo(planetIds.get(0));
        _getRewardInfo(BigInteger.valueOf(-1));
        LOG.infoExiting();
    }

    @Test
    @Order(5)
    public void setRewardIssueTest() throws Exception {
        LOG.infoEntering("setRewardIssueTest");
        KeyWallet EOAWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();
        _startRewardIssue(EOAWallet, termPeriod, false);
        BigInteger reward = _startRewardIssue(governorWallet, termPeriod, true);
        if(_getHeight().compareTo(reward) < 0)
            _startRewardIssue(governorWallet, termPeriod, true); // 리워드가 시작되기 전 호출은 성공해야함.
        else
            LOG.info("reward already started. ignore continuous call test.");

        _waitUtil(reward);

        _startRewardIssue(governorWallet, termPeriod, false); // 리워드가 시작되면 실패해야 한다고 함.
        LOG.infoExiting();
    }

    @Test
    @Order(6)
    public void reportPlanetWorkTest() throws Exception {
        LOG.infoEntering("reportPlanetWorkTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        KeyWallet EOAWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PULBIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        _startRewardIssue(governorWallet, termPeriod.multiply(BigInteger.TEN), true);

        _reportPlanetWork(EOAWallet, planetIds.get(0), false);
        _reportPlanetWork(planetManagerWallet, BigInteger.valueOf(-1), false);
        _reportPlanetWork(planetManagerWallet, planetIds.get(0), false);

        BigInteger reward = _startRewardIssue(governorWallet, termPeriod, true);
        _waitUtil(reward);
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
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PULBIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        BigInteger reward = _startRewardIssue(governorWallet, termPeriod, true);
        _waitUtil(reward);
        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        _getRewardInfo(planetIds.get(0));
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true);

        // mint second planet nft
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PULBIC);
        planetIds = _tokenIdsOf(planetWallet.getAddress(), 2, BigInteger.TWO);

        _waitUtil(_getHeight().add(termPeriod));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        _reportPlanetWork(planetManagerWallet, planetIds.get(1), true);

        _getRewardInfo(planetIds.get(0));
        _getRewardInfo(planetIds.get(1));
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0), planetIds.get(1)}, true);

        LOG.infoExiting();
    }

    @Test
    @Order(8)
    public void claimPrivatePlanetRewardTest() throws Exception {
        LOG.infoEntering("claimPrivatePlanetRewardTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();
//        int privateLockup = _getPrivateLockup().intValue();
//        int privateReleaseCycle = _getPrivateReleaseCycle().intValue();
        int privateLockup = 8;
        BigInteger privateReleaseCycle = BigInteger.valueOf(4);
        LOG.info("termPeriod : " + termPeriod);
        LOG.info("privateLockup : " + privateLockup);
        LOG.info("privateReleaseCycle : " + privateReleaseCycle);

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_PRIVATE);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        BigInteger rewardHeight = _startRewardIssue(governorWallet, termPeriod, true);
        _waitUtil(rewardHeight);

        int count = 24; // 임시값
        for(int i = 0; i < count; i++) {
            BigInteger curHeight = _getHeight();
            LOG.info("term : " + curHeight.subtract(rewardHeight).divide(termPeriod));
            _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
            if(_getRewardInfo(planetIds.get(0)) > 0) {
                _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true);
            }
            _waitUtil(curHeight.add(termPeriod));
        }

        LOG.infoExiting();
    }

    @Test
    @Order(9)
    public void claimCompanyPlanetRewardTest() throws Exception {
        LOG.infoEntering("claimPrivatePlanetRewardTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();
        LOG.info("termPeriod : " + termPeriod);

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANETTYPE_COMPANY);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        BigInteger rewardHeight = _startRewardIssue(governorWallet, termPeriod, true);
        _waitUtil(rewardHeight);

        LOG.info("ecosystem balance (before report) : " + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS));
        LOG.info("sustainable fund balance (before report) : " + txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS));
        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        LOG.info("ecosystem balance (after report) : " + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS));
        LOG.info("sustainable fund balance (after report) : " + txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS));
        _getRewardInfo(planetIds.get(0));
        LOG.info("ecosystem balance (before claim) : " + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS));
        LOG.info("sustainable fund balance (before claim) : " + txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS));
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true);
        LOG.info("ecosystem balance (after claim) : " + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS));
        LOG.info("sustainable fund balance (after claim) : " + txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS));
        _getRewardInfo(planetIds.get(0));

        LOG.infoExiting();
    }
}
