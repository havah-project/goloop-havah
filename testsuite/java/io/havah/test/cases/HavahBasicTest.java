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
import static io.havah.test.score.PlanetNFTScore.PLANET_PRIVATE;
import static io.havah.test.score.PlanetNFTScore.PLANET_PUBLIC;
import static io.havah.test.score.PlanetNFTScore.PLANET_COMPANY;
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

    public static TransactionResult  _checkAndClaimPlanetReward(Wallet wallet, BigInteger[] planetIds, boolean success, BigInteger expected, int compare) throws Exception {
        LOG.infoEntering("_checkAndClaimPlanetReward");
        BigInteger before = txHandler.getBalance(wallet.getAddress());
        LOG.info("planet balance (before claim) : " + before);
        TransactionResult result = _claimPlanetReward(wallet, planetIds, success);
        BigInteger after = txHandler.getBalance(wallet.getAddress());
        LOG.info("planet balance (after claim) : " + after);
        BigInteger fee = Utils.getTxFee(result);
        LOG.info("fee : " + fee);

        assertEquals(compare, after.subtract(before).compareTo(expected.subtract(fee)), "claim reward is not expected");

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

    public static Map<String, Object> _getRewardInfo(BigInteger planetId) throws Exception {
        try {
            RpcObject obj = chainScore.getRewardInfo(planetId);
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

    public static BigInteger _getCurrentPrivateReward(BigInteger planetStart, BigInteger total) throws IOException {
        try {
            RpcObject obj = chainScore.getIssueInfo();
            BigInteger termPeriod = obj.getItem("termPeriod").asInteger();
            BigInteger height = obj.getItem("height").asInteger();

            BigInteger privateLockup = Constants.PRIVATE_LOCKUP;
            BigInteger privateReleaseCycle = Constants.PRIVATE_RELEASE_CYCLE;

            BigInteger lockupTerm = height.subtract(planetStart).subtract(BigInteger.ONE).divide(termPeriod);

            if (lockupTerm.compareTo(privateLockup) < 0)
                return BigInteger.ZERO;

            BigInteger releaseCycle = lockupTerm.subtract(privateLockup).divide(privateReleaseCycle).add(BigInteger.ONE);
            BigInteger reward = total;
            LOG.info("releaseCycle : " + releaseCycle);
            if (releaseCycle.compareTo(BigInteger.valueOf(25)) < 0)  {
                reward = total.subtract(total.multiply(BigInteger.valueOf(24).subtract(releaseCycle)).divide(BigInteger.valueOf(24)));
            }
            return reward;
        } catch (NullPointerException e) {
            LOG.info("getIssueInfo has null infomation.");
            return BigInteger.ZERO;
        }
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
        assertEquals(true, BigInteger.valueOf(4).compareTo((BigInteger) info.get("termPeriod")) == 0);
        assertEquals(true, BigInteger.valueOf(16384).compareTo((BigInteger) info.get("issueReductionCycle")) == 0);
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
            Map<String, Object> info = _getRewardInfo(planetIds.get(0));
            assertEquals(true, BigInteger.ZERO.compareTo((BigInteger) info.get("total")) == 0);
            assertEquals(true, BigInteger.ZERO.compareTo((BigInteger) info.get("remain")) == 0);
            assertEquals(true, BigInteger.ZERO.compareTo((BigInteger) info.get("claimable")) == 0);
        }
        Map<String, Object> info = _getRewardInfo(BigInteger.valueOf(-1));
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
            Map<String, Object> info = _getRewardInfo(planetIds.get(0));
            assertEquals(true, BigInteger.ZERO.compareTo((BigInteger) info.get("total")) == 0);
        }

        Utils.waitUtilNextTerm();

        for(int i=0;i<types.length;i++) {
            List<BigInteger> planetIds = _tokenIdsOf(wallets[i].getAddress(), 1, BigInteger.ONE);
            _reportPlanetWork(EOAWallet, planetIds.get(0), false);
            _reportPlanetWork(planetManagerWallet, BigInteger.valueOf(-1), false);

            Map<String, Object> info = _getRewardInfo(planetIds.get(0));
            BigInteger before = (BigInteger) info.get("total");
            assertEquals(true, BigInteger.ZERO.compareTo((BigInteger) info.get("total")) == 0);

            _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);

            info = _getRewardInfo(planetIds.get(0));
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

        Utils.waitUtilNextTerm();
        _getPlanetInfo(planetIds.get(0));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);

        BigInteger claimable = (BigInteger) _getRewardInfo(planetIds.get(0)).get("claimable");
        BigInteger expected = _getCurrentPublicReward();

        assertEquals(0, claimable.compareTo(expected), "claimable is not expected");
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, expected, 0);

        // mint second planet nft
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANET_PUBLIC);
        planetIds = _tokenIdsOf(planetWallet.getAddress(), 2, BigInteger.TWO);

        Utils.waitUtilNextTerm();

        for(int i=0; i<planetIds.size(); i++) {
            _reportPlanetWork(planetManagerWallet, planetIds.get(i), true);
            claimable = (BigInteger) _getRewardInfo(planetIds.get(i)).get("claimable");
            expected = _getCurrentPublicReward();

            assertEquals(true, claimable.compareTo(expected) == 0, "claimable is not expected");
        }
        expected = _getCurrentPublicReward().multiply(BigInteger.TWO);
        BigInteger[] planetIdArray = new BigInteger[planetIds.size()];
        planetIds.toArray(planetIdArray);
        _checkAndClaimPlanetReward(planetWallet, planetIdArray, true, expected, 0);

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

        Utils.waitUtilNextTerm();

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        BigInteger claimable = (BigInteger) _getRewardInfo(planetIds.get(0)).get("claimable");
        BigInteger reward = _getCurrentPublicReward();
        BigInteger expectedPlanet = reward.multiply(BigInteger.valueOf(4)).divide(BigInteger.TEN);
        BigInteger expectedEco = reward.multiply(BigInteger.valueOf(6)).divide(BigInteger.TEN);
        LOG.info("reward = " + reward);
        LOG.info("Planet = " + expectedPlanet);
        LOG.info("Eco = " + expectedEco);

        assertEquals(true, claimable.compareTo(expectedPlanet) == 0, "claimable is not expected");

        BigInteger beforeEco = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
        LOG.info("ecosystem balance (before claim) : " + beforeEco);
        TransactionResult result = _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, expectedPlanet, 0);

        Utils.waitUtilNextTerm();

        BigInteger afterEco = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
        LOG.info("ecosystem balance (after claim) : " + afterEco);
        assertEquals(true, afterEco.subtract(beforeEco).compareTo(expectedEco.subtract(Utils.getTxFee(result))) >= 0, "ecosystem reward is not expected");

        LOG.infoExiting();
    }

    @Test
    public void claimPrivatePlanetRewardTest() throws Exception {
        LOG.infoEntering("claimPrivatePlanetRewardTest");
        KeyWallet planetManagerWallet = KeyWallet.create();
        KeyWallet planetWallet = KeyWallet.create();
        BigInteger termPeriod = _getTermPeriod();
        BigInteger privateLockup = Constants.PRIVATE_LOCKUP;
        BigInteger privateReleaseCycle = Constants.PRIVATE_RELEASE_CYCLE;

        LOG.info("termPeriod : " + termPeriod);
        LOG.info("privateLockup : " + privateLockup);
        LOG.info("privateReleaseCycle : " + privateReleaseCycle);

        Utils.distributeCoin(new Wallet[] {planetManagerWallet, planetWallet});

        _checkPlanetManager(governorWallet, planetManagerWallet.getAddress(), true);
        _checkAndMintPlanetNFT(planetWallet.getAddress(), PLANET_PRIVATE);
        List<BigInteger> planetIds = _tokenIdsOf(planetWallet.getAddress(), 1, BigInteger.ONE);
        BigInteger planetId = planetIds.get(0);
        BigInteger planetHeight = (BigInteger) _getPlanetInfo(planetId).get("height");

        Utils.waitUtilNextTerm();

        _reportPlanetWork(planetManagerWallet, planetId, true);
        BigInteger totalReward = _getCurrentPublicReward();
        BigInteger claimedReward = BigInteger.ZERO;

        var lockupHeight = _getStartHeightOfTerm(privateLockup, termPeriod, planetHeight.add(BigInteger.ONE));
        LOG.info("lockupHeight = " + lockupHeight);
        Utils.waitUtil(lockupHeight);

        int testTermCycle = 24;
        for (int i = 0; i < testTermCycle; i++) {
            var nextCycle = lockupHeight.add(termPeriod.multiply(privateReleaseCycle).multiply(BigInteger.valueOf(i + 1)));
            BigInteger claimable = (BigInteger) _getRewardInfo(planetId).get("claimable");
            BigInteger expected = _getCurrentPrivateReward(planetHeight, totalReward).subtract(claimedReward);
            LOG.info("claimable = " + claimable);
            LOG.info("expected = " + expected);
            assertEquals(claimable.compareTo(expected), 0, "private reward is not expected");
            _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, expected, 0);
            claimedReward = claimedReward.add(claimable);

            _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
            totalReward = totalReward.add(_getCurrentPublicReward());

            Utils.waitUtil(nextCycle);
        }
        BigInteger claimable = (BigInteger) _getRewardInfo(planetId).get("claimable");
        BigInteger expected = totalReward.subtract(claimedReward);
        LOG.info("last claim!");
        LOG.info("claimable = " + claimable);
        LOG.info("expected = " + expected);
        assertEquals(claimable.compareTo(expected), 0, "last reward is not expected");
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, expected, 0);

        LOG.infoExiting();
    }
}
