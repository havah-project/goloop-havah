package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Block;
import foundation.icon.icx.data.Bytes;
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
import score.Context;

import javax.imageio.IIOException;
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
    private static KeyWallet[] testWallets;
    private static final int testWalletNum = 3;
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

        try {
            Bytes txHash = txHandler.transfer(chain.godWallet, governorWallet.getAddress(), ICX);
            assertSuccess(txHandler.getResult(txHash));

            testWallets = new KeyWallet[testWalletNum];
            for (int i = 0; i < testWalletNum; i++) {
                testWallets[i] = KeyWallet.create();
            }
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
        assertFailure(govScore.addPlanetManager(testWallets[0], address));
        assertSuccess(govScore.addPlanetManager(governorWallet, address));
        LOG.infoExiting();

        LOG.infoEntering("isPlanetManager");
        assertEquals(chainScore.isPlanetManager(address), true);
        LOG.infoExiting();
    }

    public void _mintPlanetNFT(Wallet wallet, Address to, int type, boolean success) throws Exception {
        Bytes txHash = planetNFTScore.mintPlanet(wallet, to, type, BigInteger.ONE, BigInteger.ONE);
        if (success) {
            assertSuccess(planetNFTScore.getResult(txHash));
        } else {
            assertFailure(planetNFTScore.getResult(txHash));
        }
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

    public void _startRewardIssue() throws IOException, ResultTimeoutException {
        LOG.infoEntering("startRewardIssue");
        var height = _getHeight();
        var reward = height.add(BigInteger.valueOf(10));
        LOG.info("cur height : " + height);
        LOG.info("reward height : " + reward);
        assertFailure(govScore.startRewardIssue(testWallets[0], reward));
        assertSuccess(govScore.startRewardIssue(governorWallet, reward));
        LOG.infoExiting();
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
        if(success) {
            assertSuccess(chainScore.reportPlanetWork(wallet, planetId));
        } else {
            assertFailure(chainScore.reportPlanetWork(wallet, planetId));
        }
        LOG.infoExiting();
    }

    public void _checkAndClaimPlanetReward(Wallet wallet, BigInteger planetId) throws Exception {
        LOG.infoEntering("_checkAndClaimPlanetReward");
        LOG.info("planet balance (before claim) : " + txHandler.getBalance(wallet.getAddress()).intValue());
        _claimPlanetReward(wallet, planetId);
        LOG.info("planet balance (after claim) : " + txHandler.getBalance(wallet.getAddress()).intValue());
        LOG.infoExiting();
    }

    public void _claimPlanetReward(Wallet wallet, BigInteger planetId) throws Exception {
        assertFailure(chainScore.claimPlanetReward(testWallets[0], new BigInteger[] { planetId }));
        assertSuccess(chainScore.claimPlanetReward(wallet, new BigInteger[] { planetId }));
    }

    public void _getPlanetInfo(BigInteger planetId) throws IOException {
        LOG.infoEntering("_getPlanetInfo");
        try {
            RpcObject obj = chainScore.getPlanetInfo(planetId);
            LOG.info("owner : " + obj.getItem("owner").asAddress());
            LOG.info("usdtPrice : " + obj.getItem("usdtPrice").asInteger());
            LOG.info("havahPrice : " + obj.getItem("havahPrice").asInteger());
            LOG.info("isCompany : " + obj.getItem("isCompany").asBoolean());
            LOG.info("isPrivate : " + obj.getItem("isPrivate").asBoolean());
            LOG.info("height : " + obj.getItem("height").asInteger());
        } catch (RpcError e) {
            assertEquals(-30032, e.getCode());
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
        LOG.infoEntering("_getRewardInfo");
        RpcObject obj = chainScore.getRewardInfo(planetId);
        LOG.infoExiting();
    }

//    @Test
    public void test() throws Exception {
        LOG.infoEntering("test");
        _checkPlanetManager(testWallets[1]);
        _checkAndMintPlanetNFT(testWallets[2].getAddress());
        BigInteger planetId = _tokenIdsOf(testWallets[2].getAddress());
        _startRewardIssue();
        //TODO: getrewardinfo
        _reportPlanetWork(governorWallet, planetId, false);
        LOG.info("cur Height" + _getHeight());
        LOG.info("waiting 11 sec....");
        Thread.sleep(11000);
        LOG.info("cur Height" + _getHeight());
        //TODO: getrewardinfo
        _reportPlanetWork(governorWallet, planetId, true);
        _reportPlanetWork(governorWallet, planetId, false);
        //TODO: getrewardinfo
        _checkAndClaimPlanetReward(testWallets[2], planetId);
        LOG.infoExiting();
    }

//    @Test
    public void getPlanetInfoTest() throws Exception {
        LOG.infoEntering("getPlanetInfoTest");
        _checkPlanetManager(testWallets[1]);
        _checkAndMintPlanetNFT(testWallets[2].getAddress());
        BigInteger planetId = _tokenIdsOf(testWallets[2].getAddress());
        LOG.info("planetId : " + planetId);
        _getPlanetInfo(planetId);
        _getPlanetInfo(BigInteger.valueOf(-1));
        LOG.infoExiting();
    }

    @Test
    public void getIssueInfoTest() throws Exception {
        LOG.infoEntering("getIssueInfoTest");
        _getIssueInfo();
        LOG.infoExiting();
    }

//    @Test
    public void getRewardInfoTest() throws Exception {
        LOG.infoEntering("getRewardInfoTest");
        _checkPlanetManager(testWallets[1]);
        _checkAndMintPlanetNFT(testWallets[2].getAddress());
        BigInteger planetId = _tokenIdsOf(testWallets[2].getAddress());
        LOG.info("planetId : " + planetId);
//        _getRewardInfo(planetId);
//        _startRewardIssue();
        _getRewardInfo(BigInteger.valueOf(-1));
        LOG.infoExiting();
    }
}
