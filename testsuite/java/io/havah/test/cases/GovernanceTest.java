package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.http.HttpProvider;
import foundation.icon.test.common.Env;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.score.ChainScore;
import io.havah.test.score.GovScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Tag;

import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.fail;

//@Tag(Constants.TAG_HAVAH)
public class GovernanceTest extends TestBase {
    /*
    1. setUSDT & getUSDT
    2. startRewardIssue
        // startRewardIssue 이전 reportPlanetWork, claimPlanetReward는 의미없음.
        // startRewardIssue 이후 reportPlanetWork, claimPlanetReward가 정상동작.
    3. addPlanetManger
    4. removePlanetManager
    5. reportPlanetWork
        4 - 1. deploy planetNft
        4 - 2. mint planetNft to A
        4 - 3. reportPlanetWork with B - failure
        4 - 4. reportPlanetWork with A - success
    6. claimPlanetReward()

    &

    test for sustainable fund
    test for hoover fund
     */
    private static TransactionHandler txHandler;

    private static ChainScore chainScore;
    private static GovScore govScore;
    private static KeyWallet[] testWallets;
    private static final int testWalletNum = 3;
    private static KeyWallet governorWallet;

    @BeforeAll
    public static void setup() {
        Env.Node node = Env.nodes[0];
        Env.Channel channel = node.channels[0];
        Env.Chain chain = channel.chain;
        IconService iconService = new IconService(new HttpProvider(channel.getAPIUrl(Env.testApiVer)));
        txHandler = new TransactionHandler(iconService, chain);
        govScore = new GovScore(txHandler);
        chainScore = new ChainScore(txHandler);
        governorWallet = txHandler.getChain().governorWallet;
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
    }

    @Test
    void testAPIS() throws Exception {
        LOG.infoEntering("setUSDTPrice", "setUSDTPrice : 200");
        var origin = chainScore.getUSDTPrice();
        BigInteger price = origin.add(BigInteger.ONE);
        assertSuccess(govScore.setUSDTPrice(governorWallet, price));
        LOG.infoExiting();

        LOG.infoEntering("getUSDTPrice");
        BigInteger changed = chainScore.getUSDTPrice();
        assertEquals(price, changed);
        assertSuccess(govScore.setUSDTPrice(governorWallet, origin));

        LOG.infoEntering("addPlanetManager");
        assertSuccess(govScore.addPlanetManager(governorWallet, testWallets[1].getAddress()));
        LOG.infoExiting();

//        LOG.infoEntering("reportPlanetWork");
//        assertSuccess(govScore.reportPlanetWork(testWallets[1], BigInteger.ONE));
//        LOG.infoExiting();

        LOG.infoEntering("removePlanetManager");
        assertSuccess(govScore.removePlanetManager(governorWallet, testWallets[1].getAddress()));
        LOG.infoExiting();

    }

    void _checkUSDTPrice(Wallet wallet, boolean expected) throws Exception {
        boolean success = wallet.equals(governorWallet);
        LOG.infoEntering("checkUSDTPrice success(" + success + ")");
        assertEquals(expected, success);
        BigInteger price = chainScore.getUSDTPrice();
        BigInteger change = price.add(BigInteger.ONE);
        TransactionResult result = govScore.setUSDTPrice(wallet, change);
        LOG.info("before(" + price + "), after(" + chainScore.getUSDTPrice());
        LOG.info("Address(" + wallet.getAddress() + "), USDTPrice result(" + result + ")");
        assertEquals(success ? 1 : 0, result.getStatus().intValue(), result.toString());
        assertEquals(success ? change : price, chainScore.getUSDTPrice());

        if (success) {
            result = govScore.setUSDTPrice(wallet, change);
            assertEquals(success ? 1 : 0, result.getStatus().intValue(), result.toString());
        }
        LOG.info("checkUSDTPrice");
    }

    @Test
    void checkUSDTPrice() throws Exception {
        Wallet tmp = testWallets[0];
        _checkUSDTPrice(tmp, false);
        _checkUSDTPrice(governorWallet, true);
    }


    void _checkPlanetManager(Wallet wallet, boolean expected) throws Exception {
        boolean success = wallet.equals(governorWallet);
        assertEquals(expected, success);
        Wallet tmp = KeyWallet.create();

        assertEquals(false, chainScore.isPlanetManager(tmp.getAddress()));
        var txResult = govScore.addPlanetManager(wallet, tmp.getAddress());
        assertEquals(success ? 1 : 0, txResult.getStatus().intValue());
        assertEquals(success, chainScore.isPlanetManager(tmp.getAddress()));
    }

    @Test
    void checkPlanetManager() throws Exception {
        Wallet tmp = testWallets[0];
        _checkPlanetManager(tmp, false);
        _checkPlanetManager(governorWallet, true);
    }

//    @Test
//    void managePlanetManager() {
//
//    }
}
