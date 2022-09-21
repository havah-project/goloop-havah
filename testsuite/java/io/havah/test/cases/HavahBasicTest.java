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
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.ChainScore;
import io.havah.test.score.GovScore;
import io.havah.test.score.PlanetNFTScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.math.BigInteger;
import java.util.List;
import java.util.Map;

import static foundation.icon.test.common.Env.LOG;
import static io.havah.test.score.PlanetNFTScore.*;
import static org.junit.jupiter.api.Assertions.*;

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
        planetNFTScore = new PlanetNFTScore(governorWallet, txHandler);

        try {
            Bytes txHash = txHandler.transfer(chain.godWallet, governorWallet.getAddress(), ICX);
            assertSuccess(txHandler.getResult(txHash));
        } catch (Exception ex) {
            fail(ex.getMessage());
        }

        var height = Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(height);
    }

    private static BigInteger _getStartHeightOfTerm(BigInteger term, BigInteger termPeriod, BigInteger rewardStartHeight) {
        return term.multiply(termPeriod).add(rewardStartHeight);
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

    public static void _reportPlanetWork(Wallet wallet, BigInteger planetId, boolean success) throws Exception {
        LOG.infoEntering("reportPlanetWork", "planetId : " + planetId);
        TransactionResult result = chainScore.reportPlanetWork(wallet, planetId);
        assertEquals(success ? Constants.STATUS_SUCCESS : Constants.STATUS_FAILURE, result.getStatus(), "failure result(" + result + ")");
        LOG.infoExiting();
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

        assertTrue(compareWithBuffer(after.subtract(before), expected.subtract(fee), compare), "claimable is not expected");

        LOG.infoExiting();
        return result;
    }

    public static TransactionResult _claimPlanetReward(Wallet wallet, BigInteger[] planetIds, boolean success) throws Exception {
        TransactionResult result = chainScore.claimPlanetReward(wallet, planetIds);
        assertEquals(success ? Constants.STATUS_SUCCESS : Constants.STATUS_FAILURE, result.getStatus(), "failure result(" + result + ")");
        return result;
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

    public static Map<String, Object> _getRewardInfoOf(BigInteger planetId) throws Exception {
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

    public static BigInteger _getTermPeriod() throws IOException {
        // termPeriod : 주기당 블록 수 (Blocks) 기본 : 43200 (하루)
        RpcObject obj = chainScore.getIssueInfo();
        return obj.getItem("termPeriod").asInteger();
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

    public static TransactionResult _setPrivateClaimableRate(Wallet wallet, BigInteger denominator, BigInteger numerator, boolean success) throws Exception {
        TransactionResult result = govScore.setPrivateClaimableRate(wallet, denominator, numerator);
        assertEquals(success ? Constants.STATUS_SUCCESS : Constants.STATUS_FAILURE, result.getStatus(), "failure result(" + result + ")");
        return result;
    }

    public static Map<String, Object> _getPrivateClaimableRate() throws IOException {
        try {
            RpcObject obj = chainScore.getPrivateClaimableRate();
            return Map.of(
                    "denominator", obj.getItem("denominator").asInteger(),
                    "numerator", obj.getItem("numerator").asInteger()
            );
        } catch (RpcError e) {
            assertEquals(Constants.RPC_ERROR_INVALID_ID, e.getCode());
            LOG.info("Expected RpcError: code=" + e.getCode() + ", msg=" + e.getMessage());
        }
        return Map.of();
    }

    public static boolean checkPrivateClaimableRate(BigInteger denominator, BigInteger numerator) throws IOException {
        Map<String, Object> claimableRate = _getPrivateClaimableRate();
        if(denominator.equals(claimableRate.get("denominator")) &&
                numerator.equals(claimableRate.get("numerator"))) return true;

        return false;
    }

    public static boolean compareWithBuffer(BigInteger left, BigInteger right, BigInteger buf) {
        return left.subtract(right).abs().compareTo(buf) < 1;
    }

    @Test
    public void addPlanetManagerTest() throws Exception {
        LOG.infoEntering("addPlanetManagerTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet EOAWallet = KeyWallet.create();

        Utils.distributeCoin(new Wallet[] {planetManagerWallet, EOAWallet});

        _checkPlanetManager(EOAWallet, planetManagerWallet.getAddress(), false);
        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        LOG.infoExiting();
    }

    @Test
    public void getPlanetInfoTest() throws Exception {
        LOG.infoEntering("getPlanetInfoTest");
        Wallet[] wallets = new KeyWallet[4];
        for(int i=0; i<wallets.length; i++) {
            wallets[i] = KeyWallet.create();
        }
        Utils.distributeCoin(wallets);

        _checkPlanetManager(governorWallet, wallets[wallets.length - 1].getAddress(), true);

        int[] types = { PLANET_PUBLIC, PLANET_PRIVATE, PLANET_COMPANY };
        boolean[] isCompany = {false, false, true};
        boolean[] isPrivate = {false, true, false};
        for(int i=0;i<types.length;i++) {
            Address address = wallets[i].getAddress();
            _checkAndMintPlanetNFT(address, types[i]);
            List<BigInteger> planetIds = _tokenIdsOf(address, 1, BigInteger.ONE);
            Map<String, Object> info = _getPlanetInfo(planetIds.get(0));
            assertEquals(true, address.equals(info.get("owner")));
            assertEquals(isCompany[i], info.get("isCompany"));
            assertEquals(isPrivate[i], info.get("isPrivate"));
        }
        Map<String, Object> info = _getPlanetInfo(BigInteger.valueOf(-1));
        assertEquals(true, info.isEmpty());
        LOG.infoExiting();
    }

    @Test
    public void getIssueInfoTest() throws Exception {
        LOG.infoEntering("getIssueInfoTest");
        Map<String, Object> info = _getIssueInfo();
        assertEquals(BigInteger.valueOf(8), info.get("termPeriod"));
        assertEquals(BigInteger.valueOf(131072), info.get("issueReductionCycle"));
        LOG.infoExiting();
    }

    @Test
    public void getRewardInfoTest() throws Exception {
        LOG.infoEntering("getRewardInfoTest");
        Wallet[] wallets = new KeyWallet[4];
        for(int i=0; i<wallets.length; i++) {
            wallets[i] = KeyWallet.create();
        }
        Utils.distributeCoin(wallets);

        _checkPlanetManager(governorWallet, wallets[wallets.length - 1].getAddress(), true);

        int[] types = { PLANET_PUBLIC, PLANET_PRIVATE, PLANET_COMPANY };
        for(int i=0;i<types.length;i++) {
            Address address = wallets[i].getAddress();
            _checkAndMintPlanetNFT(address, types[i]);
            List<BigInteger> planetIds = _tokenIdsOf(address, 1, BigInteger.ONE);
            Map<String, Object> info = _getRewardInfoOf(planetIds.get(0));
            assertEquals(BigInteger.ZERO, info.get("total"));
            assertEquals(BigInteger.ZERO, info.get("remain"));
            assertEquals(BigInteger.ZERO, info.get("claimable"));
        }
        Map<String, Object> info = _getRewardInfoOf(BigInteger.valueOf(-1));
        assertEquals(true, info.isEmpty());

        LOG.infoExiting();
    }

    @Test
    public void reportPlanetWorkTest() throws Exception {
        LOG.infoEntering("reportPlanetWorkTest");
        Wallet[] wallets = new KeyWallet[5];
        for(int i=0; i<wallets.length; i++) {
            wallets[i] = KeyWallet.create();
        }
        Utils.distributeCoin(wallets);
        Wallet planetManagerWallet = wallets[wallets.length - 1];
        Wallet EOAWallet = wallets[wallets.length - 2];

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);

        int[] types = { PLANET_PUBLIC, PLANET_PRIVATE, PLANET_COMPANY };
        for(int i=0;i<types.length;i++) {
            _checkAndMintPlanetNFT(wallets[i].getAddress(), types[i]);

            List<BigInteger> planetIds = _tokenIdsOf(wallets[i].getAddress(), 1, BigInteger.ONE);
            _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
            Map<String, Object> info = _getRewardInfoOf(planetIds.get(0));
            assertEquals(BigInteger.ZERO, info.get("total"));
        }

        Utils.waitUntilNextTerm();

        for(int i=0;i<types.length;i++) {
            List<BigInteger> planetIds = _tokenIdsOf(wallets[i].getAddress(), 1, BigInteger.ONE);
            _reportPlanetWork(EOAWallet, planetIds.get(0), false);
            _reportPlanetWork(planetManagerWallet, BigInteger.valueOf(-1), false);

            Map<String, Object> info = _getRewardInfoOf(planetIds.get(0));
            BigInteger before = (BigInteger) info.get("total");
            assertEquals(BigInteger.ZERO, info.get("total"));

            _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);

            info = _getRewardInfoOf(planetIds.get(0));
            assertEquals(true, before.compareTo((BigInteger) info.get("total")) < 0);
        }
        LOG.infoExiting();
    }

    @Test
    public void claimPublicPlanetRewardTest() throws Exception {
        LOG.infoEntering("claimPublicPlanetRewardTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();

        Utils.distributeCoin(new Wallet[] {planetManagerWallet, planetWallet});

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANET_PUBLIC);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        Utils.waitUntilNextTerm();
        _getPlanetInfo(planetIds.get(0));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);

        BigInteger claimable = (BigInteger) _getRewardInfoOf(planetIds.get(0)).get("claimable");
        BigInteger expected = _getCurrentPublicReward();

        assertTrue(compareWithBuffer(claimable, expected, BigInteger.TWO), "claimable is not expected");
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, expected, BigInteger.TWO);

        // mint second planet nft
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANET_PUBLIC);
        planetIds = _tokenIdsOf(planetWallet.getAddress(), 2, BigInteger.TWO);

        Utils.waitUntilNextTerm();

        for(int i=0; i<planetIds.size(); i++) {
            _reportPlanetWork(planetManagerWallet, planetIds.get(i), true);
            claimable = (BigInteger) _getRewardInfoOf(planetIds.get(i)).get("claimable");
            expected = _getCurrentPublicReward();

            assertTrue(compareWithBuffer(claimable, expected, BigInteger.TWO), "claimable is not expected");
        }
        expected = _getCurrentPublicReward().multiply(BigInteger.TWO);
        BigInteger[] planetIdArray = new BigInteger[planetIds.size()];
        planetIds.toArray(planetIdArray);
        _checkAndClaimPlanetReward(planetWallet, planetIdArray, true, expected, BigInteger.TWO);

        LOG.infoExiting();
    }

    @Test
    public void claimCompanyPlanetRewardTest() throws Exception {
        LOG.infoEntering("claimCompanyPlanetRewardTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();

        Utils.distributeCoin(new Wallet[] {planetManagerWallet, planetWallet});

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANET_COMPANY);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);

        _getPlanetInfo(planetIds.get(0));

        Utils.waitUntilNextTerm();

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        BigInteger claimable = (BigInteger) _getRewardInfoOf(planetIds.get(0)).get("claimable");
        BigInteger reward = _getCurrentPublicReward();
        BigInteger expectedPlanet = reward.multiply(BigInteger.valueOf(4)).divide(BigInteger.TEN);
        BigInteger expectedEco = reward.multiply(BigInteger.valueOf(6)).divide(BigInteger.TEN);
        LOG.info("reward = " + reward);
        LOG.info("Planet = " + expectedPlanet);
        LOG.info("Eco = " + expectedEco);

        assertTrue(compareWithBuffer(claimable, expectedPlanet, BigInteger.TWO), "claimable is not expected");

        BigInteger beforeEco = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
        LOG.info("ecosystem balance (before claim) : " + beforeEco);
        TransactionResult result = _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, expectedPlanet, BigInteger.TWO);

        Utils.waitUntilNextTerm();
        Utils.waitUntilNextTerm();

        BigInteger afterEco = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
        LOG.info("ecosystem balance (after claim) : " + afterEco);
        assertEquals(true, afterEco.subtract(beforeEco).compareTo(expectedEco.subtract(Utils.getTxFee(result))) >= 0, "ecosystem reward is not expected");

        LOG.infoExiting();
    }

    @Test
    public void claimPrivatePlanetRewardTest() throws Exception {
        LOG.infoEntering("claimPrivatePlanetRewardTest");
        // setup
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();

        Utils.distributeCoin(new Wallet[] {planetManagerWallet, planetWallet});

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANET_PRIVATE);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);
        BigInteger planetId = planetIds.get(0);

        Utils.waitUntilNextTerm();

        // report planet work 2 times
        BigInteger totalReward = BigInteger.ZERO;
        for (int i = 0; i < 2; i++) {
            _reportPlanetWork(planetManagerWallet, planetId, true);
            totalReward = totalReward.add(_getCurrentPublicReward());
            Utils.waitUntilNextTerm();
        }

        // check getPrivateClaimableRate default value
        assertTrue(checkPrivateClaimableRate(BigInteger.valueOf(24), BigInteger.ZERO));

        // check setPrivateClaimableRate param
        _setPrivateClaimableRate(planetWallet, BigInteger.valueOf(1000), BigInteger.ONE, false);
        _setPrivateClaimableRate(governorWallet, BigInteger.ZERO, BigInteger.ONE, false);
        _setPrivateClaimableRate(governorWallet, BigInteger.valueOf(-1), BigInteger.ONE, false);
        _setPrivateClaimableRate(governorWallet, BigInteger.ONE, BigInteger.valueOf(-1), false);
        _setPrivateClaimableRate(governorWallet, BigInteger.valueOf(1000), BigInteger.valueOf(11115), false);

        BigInteger denominator = BigInteger.valueOf(100);
        BigInteger numerator = BigInteger.ONE;
        _setPrivateClaimableRate(governorWallet, denominator, numerator, true);
        assertTrue(checkPrivateClaimableRate(denominator, numerator));

        Utils.waitUntilNextTerm();

        // check reward is expected
        Map<String, Object> info = _getRewardInfoOf(planetId);
        BigInteger expected = totalReward.divide(denominator);
        assertEquals(true, compareWithBuffer(expected, (BigInteger) info.get("claimable"), BigInteger.TWO));

        // change private claimable rate
        numerator = BigInteger.TEN;
        _setPrivateClaimableRate(governorWallet, denominator, numerator, true);
        assertTrue(checkPrivateClaimableRate(denominator, numerator));

        Utils.waitUntilNextTerm();

        // check reward is expected
        expected = totalReward.multiply(numerator).divide(denominator);
        info = _getRewardInfoOf(planetId);
        BigInteger claimable = (BigInteger) info.get("claimable");
        assertEquals(true, compareWithBuffer(expected, claimable, BigInteger.TWO));

        // claim reward
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[] {planetId}, true, expected, BigInteger.TWO);
        BigInteger claimedReward = claimable;

        info = _getRewardInfoOf(planetId);
        assertEquals(BigInteger.ZERO, info.get("claimable"));
        assertEquals(true, compareWithBuffer(totalReward.subtract(claimedReward), (BigInteger) info.get("remain"), BigInteger.TWO));

        Utils.waitUntilNextTerm();

        // change private claimable rate
        numerator = BigInteger.valueOf(15);
        _setPrivateClaimableRate(governorWallet, denominator, numerator, true);
        assertTrue(checkPrivateClaimableRate(denominator, numerator));

        // check claimable is expected
        expected = totalReward.multiply(numerator).divide(denominator).subtract(claimedReward);
        claimable = (BigInteger) _getRewardInfoOf(planetId).get("claimable");
        assertTrue(compareWithBuffer(expected, claimable, BigInteger.TWO));

        // reportPlanetWork
        _reportPlanetWork(planetManagerWallet, planetId, true);
        totalReward = totalReward.add(_getCurrentPublicReward());
        Utils.waitUntilNextTerm();

        // check claimable is expected
        expected = totalReward.multiply(numerator).divide(denominator).subtract(claimedReward);
        claimable = (BigInteger) _getRewardInfoOf(planetId).get("claimable");
        assertTrue(compareWithBuffer(expected, claimable, BigInteger.TWO));

        LOG.infoExiting();
    }

    @Test
    void rewardInfo() throws Exception {
        Wallet holder = KeyWallet.create();
        for (var type : new int[]{PLANET_PUBLIC, PLANET_PRIVATE, PLANET_COMPANY}) {
            _mintPlanetNFT(governorWallet, holder.getAddress(), type, true);
        }

        Utils.waitUntilNextTerm();
        Utils.waitUtil(Utils.getHeightNext(1));

        var obj = chainScore.getRewardInfo();
        var num = planetNFTScore.totalSupply();
        var expected = Constants.INITIAL_ISSUE_AMOUNT.divide(num);
        var reward = obj.getItem("rewardPerActivePlanet").asInteger();
        assertTrue(compareWithBuffer(reward, expected, BigInteger.TWO));

        for (var type : new int[]{PLANET_PUBLIC, PLANET_PRIVATE, PLANET_COMPANY}) {
            _mintPlanetNFT(governorWallet, holder.getAddress(), type, true);
        }
        Utils.waitUntilNextTerm();
        Utils.waitUtil(Utils.getHeightNext(1));

        num = planetNFTScore.totalSupply();
        expected = Constants.INITIAL_ISSUE_AMOUNT.divide(num);
        obj = chainScore.getRewardInfo();
        reward = obj.getItem("rewardPerActivePlanet").asInteger();
        assertTrue(compareWithBuffer(reward, expected, BigInteger.TWO));
    }
}