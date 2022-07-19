package io.havah.test.cases;

import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.EcosystemScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;

@Tag(Constants.TAG_HAVAH)
public class EcosystemTest extends TestBase {
    private static TransactionHandler txHandler;
    private static EcosystemScore ecoScore;
    private static Wallet ecoOwner;
    private static Wallet[] wallets;

    @BeforeAll
    static void setup() throws Exception {
        txHandler = Utils.getTxHandler();
        ecoScore = new EcosystemScore(txHandler);
        ecoOwner = Utils.getGovernor();
        wallets = new Wallet[3];
        wallets[0] = ecoOwner;
        for (int i = 1; i < wallets.length; i++) {
            wallets[i] = KeyWallet.create();
        }
        Utils.distributeCoin(wallets);
    }

    private void _transferExceedAndMax(Address receiver, BigInteger lockedAmount) throws Exception {
        // test transfer exceed amount
        LOG.info("_transferExceedAndMax balance(" + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS) + "), amount(" + lockedAmount + ")");
        var maxAmount = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS).subtract(lockedAmount);
        if (maxAmount.equals(BigInteger.ZERO)) {
            LOG.info("No transferable HVH. balance(" + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS) + ", locked(" + lockedAmount + ")");
            return;
        }
        var result = txHandler.getResult(ecoScore.transfer(ecoOwner, receiver, maxAmount.add(BigInteger.ONE)));
        assertEquals(BigInteger.ZERO, result.getStatus(),
                String.format("balance(%s), locked(%s)", txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS), lockedAmount));

        // test transfer all transferable
        LOG.info("transfer " + maxAmount + " to " + receiver);
        result = txHandler.getResult(ecoScore.transfer(ecoOwner, receiver, maxAmount));
        assertSuccess(result);
        assertEquals(BigInteger.ONE, result.getStatus());
    }

    @Test
    public void checkLockupSchedule() throws Exception {
        var schedules = ecoScore.getLockupSchedule();
        var lockedAmount = Constants.ECOSYSTEM_INITIAL_BALANCE;
        Address receiver = wallets[2].getAddress();
        Wallet noPermission = wallets[1];
        final BigInteger withdrawAmount = BigInteger.valueOf(1);
        // check height and withdraw
        for (var schedule : schedules) {
            var cur = Utils.getHeight();
            LOG.info("curHeight(" + cur + ")");
            // wait until schedule.blockHeight
            // cur <= schedule
            var blockHeight = schedule.getBlockHeight();
            var amount = schedule.getAmount();
            if (cur.compareTo(blockHeight) <= 0) {

                // failure case - not owner wallet
                var txHash = ecoScore.transfer(noPermission, receiver, withdrawAmount);
                assertSuccess(txHandler.getResult(txHash));

                // success case - owner wallet
                var ecoBalance = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
                txHash = ecoScore.transfer(ecoOwner, receiver, withdrawAmount);
                var result = txHandler.getResult(txHash);
                boolean available = ecoBalance.subtract(lockedAmount).compareTo(withdrawAmount) >= 0;
                assertEquals(available ? BigInteger.ONE : BigInteger.ZERO, result.getStatus(),
                        "balance(" + ecoBalance + "), locked(" + lockedAmount + "), available(" + available + ")");
                _transferExceedAndMax(receiver, lockedAmount);
                Utils.waitUtil(blockHeight);
                lockedAmount = amount;
                LOG.info("lock amount(" + amount + ") on (" + blockHeight + ") height");
            } else {
                lockedAmount = BigInteger.ZERO;
            }
        }
        _transferExceedAndMax(receiver, lockedAmount);
    }
}
