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
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.math.BigInteger;
import java.util.List;
import java.util.Map;

import static foundation.icon.test.common.Env.LOG;
import static io.havah.test.cases.HavahBasicTest.*;
import static io.havah.test.score.PlanetNFTScore.PLANET_PUBLIC;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.fail;

@Tag(Constants.TAG_HAVAH_EXTRA)
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

    @Test
    public void setRewardIssueTest() throws Exception {
        LOG.infoEntering("setRewardIssueTest");
        KeyWallet EOAWallet = KeyWallet.create();
        BigInteger startReward = BigInteger.valueOf(5);
        _startRewardIssue(EOAWallet, startReward, false);
        BigInteger reward = _startRewardIssue(governorWallet, startReward, true);
        if(Utils.getHeight().compareTo(reward) < 0)
            _startRewardIssue(governorWallet, startReward, true); // 리워드가 시작되기 전 호출은 성공해야함.
        else
            LOG.info("reward already started. ignore continuous call test.");

        Utils.waitUtil(reward);

        _startRewardIssue(governorWallet, startReward, false); // 리워드가 시작되면 실패해야 한다고 함.
        LOG.infoExiting();
    }

    @Test
    public void issueReductionCycleTest() throws Exception {
        LOG.infoEntering("issueReductionCycleTest");
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
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, termReward, 0);

        Utils.waitUtil(((BigInteger)_getIssueInfo().get("issueStart")).add(termPeriod.multiply(issueReductionCycle)));

        _reportPlanetWork(planetManagerWallet, planetIds.get(0), true);
        claimable = (BigInteger) _getRewardInfo(planetIds.get(0)).get("claimable");
        termReward = _getCurrentPublicReward();
        LOG.info("reduction termReward : " + termReward);
        assertEquals(claimable.compareTo(termReward), 0, "reduction term reward is not equals to claimable");
        _checkAndClaimPlanetReward(planetWallet, new BigInteger[]{planetIds.get(0)}, true, termReward, 0);

        LOG.infoExiting();
    }
}
