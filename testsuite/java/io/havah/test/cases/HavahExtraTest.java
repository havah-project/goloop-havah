package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
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
import io.havah.test.common.Utils;
import io.havah.test.score.ChainScore;
import io.havah.test.score.GovScore;
import io.havah.test.score.PlanetNFTScore;
import org.junit.jupiter.api.*;

import java.io.IOException;
import java.math.BigInteger;
import java.util.List;
import java.util.Map;

import static foundation.icon.test.common.Env.LOG;
import static io.havah.test.score.PlanetNFTScore.PLANET_PUBLIC;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.fail;

@Tag(Constants.TAG_HAVAH_EXTRA)
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
public class HavahExtraTest extends TestBase {
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
        planetNFTScore = new PlanetNFTScore(governorWallet, txHandler);

        try {
            Bytes txHash = txHandler.transfer(chain.godWallet, governorWallet.getAddress(), ICX);
            assertSuccess(txHandler.getResult(txHash));
        } catch (Exception ex) {
            fail(ex.getMessage());
        }
    }

    public static BigInteger _getTermPeriod() throws IOException {
        // termPeriod : 주기당 블록 수 (Blocks) 기본 : 43200 (하루)
        RpcObject obj = chainScore.getIssueInfo();
        return obj.getItem("termPeriod").asInteger();
    }

    private static BigInteger _startRewardIssue(Wallet wallet, BigInteger addHeight, boolean success) throws IOException, ResultTimeoutException {
        LOG.infoEntering("_startRewardIssue", "expect : " + success);
        var height = Utils.getHeight();
        var reward = height.add(addHeight);
        LOG.info("cur height : " + height);
        LOG.info("reward height : " + reward);
        TransactionResult result = govScore.startRewardIssue(wallet, reward);
        assertEquals(success ? Constants.STATUS_SUCCESS : Constants.STATUS_FAILURE, result.getStatus(), "failure result(" + result + ")");
        LOG.infoExiting();
        return reward;
    }

    private static BigInteger _getIssueReductionCycle() throws Exception {
        RpcObject obj = chainScore.getIssueInfo();
        return obj.getItem("issueReductionCycle").asInteger();
    }

    public static void _checkPlanetManager(Wallet wallet, Address address, boolean success) throws Exception {
        LOG.infoEntering("addPlanetManager");
        TransactionResult result = govScore.addPlanetManager(wallet, address);
        assertEquals(success ? Constants.STATUS_SUCCESS : Constants.STATUS_FAILURE, result.getStatus(), "failure result(" + result + ")");
        LOG.infoExiting();

        LOG.infoEntering("isPlanetManager");
        assertEquals(chainScore.isPlanetManager(address), success);
        LOG.infoExiting();
    }

    public static void _mintPlanetNFT(Wallet wallet, Address to, int type, boolean success) throws Exception {
        Bytes txHash = planetNFTScore.mintPlanet(wallet, to, type, BigInteger.ONE, BigInteger.ONE);
        TransactionResult result = planetNFTScore.getResult(txHash);
        assertEquals(success ? Constants.STATUS_SUCCESS : Constants.STATUS_FAILURE, result.getStatus(), "failure result(" + result + ")");
    }

    public static void _checkAndMintPlanetNFT(Address to, int type) throws Exception {
        LOG.infoEntering("_checkAndMintPlanetNFT", "mint PlanetNFT type : " + type);
        var oldBalance = planetNFTScore.balanceOf(to).intValue();
        var oldTotalSupply = planetNFTScore.totalSupply().intValue();
        LOG.info("PlanetNFT Balance before mint : " + oldBalance);
        LOG.info("PlanetNFT totalSupply before mint : " + oldTotalSupply);

        _mintPlanetNFT(governorWallet, to, type, true);

        // compare nft balance
        var balance = planetNFTScore.balanceOf(to).intValue();
        assertEquals(oldBalance + 1, balance);
        LOG.info("PlanetNFT Balance : after mint " + balance);

        // compare nft supply
        var totalSupply = planetNFTScore.totalSupply().intValue();
        assertEquals(oldTotalSupply + 1, totalSupply);
        LOG.info("PlanetNFT totalSupply : after mint " + totalSupply);
        LOG.infoExiting();
    }

    public static List<BigInteger> _tokenIdsOf(Address address, int idCount, BigInteger balance) throws Exception {
        LOG.infoEntering("_tokenIdsOf", "address : " + address);
        PlanetNFTScore.TokenIds ids = planetNFTScore.tokenIdsOf(address, 0, idCount);
        assertEquals(idCount, ids.tokenIds.size());
        assertEquals(balance, ids.balance);
        LOG.infoExiting();
        return ids.tokenIds;
    }

    public static Map<String, Object> _getPlanetInfo(BigInteger planetId) throws IOException {
        try {
            RpcObject obj = chainScore.getPlanetInfo(planetId);
            return Map.of(
                    "owner", obj.getItem("owner").asAddress(),
                    "usdtPrice", obj.getItem("usdtPrice").asInteger(),
                    "havahPrice", obj.getItem("havahPrice").asInteger(),
                    "isCompany", obj.getItem("isCompany").asBoolean(),
                    "isPrivate", obj.getItem("isPrivate").asBoolean(),
                    "height", obj.getItem("height").asInteger()
            );
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        return Map.of();
    }

    public static void _reportPlanetWork(Wallet wallet, BigInteger planetId, boolean success) throws Exception {
        LOG.infoEntering("reportPlanetWork", "planetId : " + planetId);
        TransactionResult result = chainScore.reportPlanetWork(wallet, planetId);
        assertEquals(success ? Constants.STATUS_SUCCESS : Constants.STATUS_FAILURE, result.getStatus(), "failure result(" + result + ")");
        LOG.infoExiting();
    }

    public static Map<String, Object> _getRewardInfo(BigInteger planetId) throws Exception {
        try {
            RpcObject obj = chainScore.getRewardInfoOf(planetId);
            return Map.of(
                    "total", obj.getItem("total").asInteger(),
                    "remain", obj.getItem("remain").asInteger(),
                    "claimable", obj.getItem("claimable").asInteger(),
                    "height", obj.getItem("height").asInteger()
            );
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        return Map.of();
    }

    public static BigInteger _getCurrentPublicReward() throws IOException {
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

    public static TransactionResult  _checkAndClaimPlanetReward(Wallet wallet, BigInteger[] planetIds, boolean success, BigInteger expected, BigInteger compare) throws Exception {
        LOG.infoEntering("_checkAndClaimPlanetReward");
        BigInteger before = txHandler.getBalance(wallet.getAddress());
        LOG.info("planet balance (before claim) : " + before);
        TransactionResult result = _claimPlanetReward(wallet, planetIds, success);
        BigInteger after = txHandler.getBalance(wallet.getAddress());
        LOG.info("planet balance (after claim) : " + after);
        BigInteger fee = Utils.getTxFee(result);
        LOG.info("fee : " + fee);

        assertEquals(true, compare.compareTo(after.subtract(before).subtract(expected.subtract(fee)).abs()) > -1, "claimable is not expected");
//        assertEquals(compare, after.subtract(before).compareTo(expected.subtract(fee)), "claim reward is not expected");

        LOG.infoExiting();
        return result;
    }

    public static TransactionResult _claimPlanetReward(Wallet wallet, BigInteger[] planetIds, boolean success) throws Exception {
        TransactionResult result = chainScore.claimPlanetReward(wallet, planetIds);
        assertEquals(success ? Constants.STATUS_SUCCESS : Constants.STATUS_FAILURE, result.getStatus(), "failure result(" + result + ")");
        return result;
    }

    public static Map<String, Object> _getIssueInfo() throws IOException {
        RpcObject obj = chainScore.getIssueInfo();
        try {
            return Map.of(
                    "termPeriod", obj.getItem("termPeriod").asInteger(),
                    "issueReductionCycle", obj.getItem("issueReductionCycle").asInteger(),
                    "height", obj.getItem("height").asInteger(),
                    "termSequence", obj.getItem("termSequence").asInteger(),
                    "issueStart", obj.getItem("issueStart").asInteger()
            );
        } catch (NullPointerException e) {
            LOG.info("startRewardIssue not called. termSequence and issueStart is null.");
        }
        return Map.of();
    }

    @Test
    @Order(1)
    public void setRewardIssueTest() throws Exception {
        LOG.infoEntering("setRewardIssueTest");
        KeyWallet EOAWallet = KeyWallet.create();

        Utils.distributeCoin(new Wallet[] { EOAWallet });

        BigInteger startReward = BigInteger.valueOf(5);
        _startRewardIssue(EOAWallet, startReward, false);
        BigInteger reward = _startRewardIssue(governorWallet, startReward, true);
        if(Utils.getHeight().compareTo(reward) < 0)
            _startRewardIssue(governorWallet, startReward, true); // 리워드가 시작되기 전 호출은 성공해야함.
        else
            LOG.info("reward already started. ignore continuous call test.");

        Utils.waitUtil(reward.add(BigInteger.ONE));

        _startRewardIssue(governorWallet, startReward, false); // 리워드가 시작되면 실패해야 한다고 함.
        LOG.infoExiting();
    }

    @Test
    @Order(2)
    public void issueReductionCycleTest() throws Exception {
        LOG.infoEntering("issueReductionCycleTest");

        var height = Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(height);

        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();
        BigInteger issueReductionCycle = _getIssueReductionCycle();

        Utils.distributeCoin(new Wallet[] {planetManagerWallet, planetWallet});

        LOG.info("termPeriod : " + termPeriod);
        LOG.info("issueReductionCycle : " + issueReductionCycle);

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANET_PUBLIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        _getPlanetInfo(planetIds.get(0));

        Utils.waitUtil(Utils.getHeight().add(termPeriod));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        BigInteger claimable = (BigInteger) _getRewardInfo(planetIds.get(0)).get("claimable");
        BigInteger termReward = _getCurrentPublicReward();
        LOG.info("termReward : " + termReward);
        assertEquals(claimable.compareTo(termReward), 0, "term reward is not equals to claimable");
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, termReward, BigInteger.TWO);

        Utils.waitUtil(((BigInteger)_getIssueInfo().get("issueStart")).add(termPeriod.multiply(issueReductionCycle)));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        claimable = (BigInteger) _getRewardInfo(planetIds.get(0)).get("claimable");
        termReward = _getCurrentPublicReward();
        LOG.info("reduction termReward : " + termReward);
        assertEquals(claimable.compareTo(termReward), 0, "reduction term reward is not equals to claimable");
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, termReward,  BigInteger.TWO);

        LOG.infoExiting();
    }
}
