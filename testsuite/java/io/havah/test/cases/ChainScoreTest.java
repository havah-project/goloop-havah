package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.GovScore;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.ChainScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.*;

@Tag(Constants.TAG_HAVAH)
public class ChainScoreTest extends TestBase {

    private static TransactionHandler txHandler;
    private static ChainScore chainScore;
    private static GovScore govScore;
    private static KeyWallet[] testWallets;
    private static final int testWalletNum = 3;
    private static KeyWallet governor;
    private static IconService iconService;

    @BeforeAll
    public static void init() throws Exception {
        iconService = Utils.getIconService();
        txHandler = Utils.getTxHandler();
        chainScore = new ChainScore(txHandler);
        govScore = new GovScore(txHandler);
        governor = txHandler.getChain().governorWallet;
        try {
            testWallets = new KeyWallet[testWalletNum];
            for (int i = 0; i < testWalletNum; i++) {
                testWallets[i] = KeyWallet.create();
            }
            Utils.distributeCoin(testWallets);
        } catch (Exception ex) {
            LOG.info("EX(" + ex + ")");
            ex.printStackTrace();
            fail(ex.getMessage());
        }
    }

    @Test
    public void setScoreOwner() throws Exception {
        // change some score owner
        var govAddr = govScore.getAddress();
        var noPermission = testWallets[0];
        var newOwner = testWallets[1];

        final String CHECK_STEP = "query";
        var queryStep = govScore.getMaxStepLimits().get(CHECK_STEP);
        var changed = queryStep.add(BigInteger.ONE);
        var result = govScore.setMaxStepLimit(CHECK_STEP, changed);
        assertSuccess(result);
        assertEquals(changed, govScore.getMaxStepLimits().get(CHECK_STEP));

        // change SCORE owner
        // failure case
        result = chainScore.setScoreOwner(noPermission, govAddr, newOwner.getAddress());
        assertFailure(result);

        // success case
        result = chainScore.setScoreOwner(governor, govAddr, newOwner.getAddress());
        assertSuccess(result);

        changed = changed.add(BigInteger.ONE);
        // call by old owner
        result = govScore.setMaxStepLimit(CHECK_STEP, changed);
        assertFailure(result);
        assertNotEquals(changed, govScore.getMaxStepLimits().get(CHECK_STEP));

        // set new owner and call
        govScore.setWallet(newOwner);
        result = govScore.setMaxStepLimit(CHECK_STEP, changed);
        assertSuccess(result);
        assertEquals(changed, govScore.getMaxStepLimits().get(CHECK_STEP));

        // revert
        result = chainScore.setScoreOwner(governor, govAddr, newOwner.getAddress());
        assertFailure(result);

        result = chainScore.setScoreOwner(newOwner, govAddr, governor.getAddress());
        assertSuccess(result);

        govScore.setWallet(governor);
        result = govScore.setMaxStepLimit(CHECK_STEP, queryStep);
        assertSuccess(result);
        assertEquals(queryStep, govScore.getMaxStepLimits().get(CHECK_STEP));
    }

    @Test
    public void burnHVH() throws Exception {
        LOG.infoEntering("BURN HVH");
        if (Utils.isRewardIssued()) {
            Utils.waitUntilNextTerm();
            Utils.waitUtil(Utils.getHeightNext(1));
        }
        var supply = iconService.getTotalSupply().execute();
        var burnedAmount = BigInteger.ONE;
        for (var wallet : testWallets) {
            var balance = txHandler.getBalance(wallet.getAddress());
            var txHash = txHandler.transfer(wallet, Constants.CHAINSCORE_ADDRESS, burnedAmount);
            var result = txHandler.getResult(txHash);
            assertSuccess(result);

            var fee = result.getStepPrice().multiply(result.getStepUsed());
            var curBalance = txHandler.getBalance(wallet.getAddress());
            assertEquals(balance.subtract(fee).subtract(burnedAmount), curBalance);

            var changedSupply = iconService.getTotalSupply().execute();
            LOG.info("burn HVH : before(" + supply + "), after(" + changedSupply + "), burned(" + burnedAmount + ")");
            assertEquals(supply.subtract(burnedAmount), changedSupply);
            supply = changedSupply;
            burnedAmount = burnedAmount.add(BigInteger.ONE);
        }
        LOG.infoExiting();
    }
}
