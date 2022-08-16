package io.havah.test.cases;

import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.test.common.TestBase;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.EcosystemScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

@Tag(Constants.TAG_HAVAH)
public class EcosystemTest extends TestBase {
    private static EcosystemScore ecoScore;
    private static Wallet ecoOwner;
    private static Wallet[] wallets;

    @BeforeAll
    static void setup() throws Exception {
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
        // check enough balance for transfer
        LOG.info("_transferExceedAndMax balance(" + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS) + "), lockedAmount(" + lockedAmount + ")");
        var maxAmount = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS).subtract(lockedAmount);
        if (maxAmount.equals(BigInteger.ZERO)) {
            LOG.info("No transferable HVH. balance(" + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS) + ", locked(" + lockedAmount + ")");
            return;
        }
        assertTrue(maxAmount.compareTo(BigInteger.ZERO) > 0);

        // test transfer exceed amount
        var exceedAmount = maxAmount.add(BigInteger.ONE);
        LOG.info("exceedAmount transfer " + exceedAmount + " to " + receiver);
        var result = txHandler.getResult(ecoScore.transfer(ecoOwner, receiver, exceedAmount));
        assertFailure(result, "balance(" + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS) + "), locked(" + lockedAmount + ")");

        // test transfer all transferable
        LOG.info("maxAmount transfer " + maxAmount + " to " + receiver);
        result = txHandler.getResult(ecoScore.transfer(ecoOwner, receiver, maxAmount));
        assertSuccess(result);
    }

    @Test
    public void checkLockupSchedule() throws Exception {
        LOG.infoEntering("checkLockupSchedule");
        var schedules = ecoScore.getLockupSchedule();
        var lockedAmount = Constants.ECOSYSTEM_INITIAL_BALANCE;
        Address receiver = wallets[2].getAddress();
        Wallet noPermission = wallets[1];
        final BigInteger withdrawAmount = BigInteger.valueOf(1);
        // check height and withdraw
        for (var schedule : schedules) {
            var cur = Utils.getHeight();
            // wait until schedule.blockHeight
            // cur <= schedule
            var blockHeight = schedule.getBlockHeight();
            var amount = schedule.getAmount();
            if (cur.compareTo(blockHeight) <= 0) {

                // failure case - not owner wallet
                var txHash = ecoScore.transfer(noPermission, receiver, withdrawAmount);
                assertFailure(txHandler.getResult(txHash));

                // success case - owner wallet
                var ecoBalance = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
                txHash = ecoScore.transfer(ecoOwner, receiver, withdrawAmount);
                boolean available = ecoBalance.subtract(lockedAmount).compareTo(withdrawAmount) >= 0;
                var result = txHandler.getResult(txHash);
                if (available) {
                    assertSuccess(result);
                } else {
                    assertFailure(result);
                }

                _transferExceedAndMax(receiver, lockedAmount);
                Utils.waitUtil(blockHeight);
                lockedAmount = amount;
                LOG.info("lock amount(" + amount + ") on (" + blockHeight + ") height");
            } else {
                lockedAmount = BigInteger.ZERO;
            }
        }
        _transferExceedAndMax(receiver, lockedAmount);
        // transfer from receiver to ecosystem
        var receiverWallet = wallets[2];
        txHandler.transfer(receiverWallet, Constants.ECOSYSTEM_ADDRESS, txHandler.getBalance(receiver).subtract(BigInteger.TEN.pow(20)));
        LOG.infoExiting();
    }

    void _transferAndCheck(Wallet wallet, Address receiver, BigInteger amount, boolean success) throws Exception {
        if (Utils.isRewardIssued()) {
            Utils.waitUtil(Utils.getHeightUntilNextTerm().add(BigInteger.ONE));
        }
        var ecoBal = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
        var receiverBal = txHandler.getBalance(receiver);
        var txHash = ecoScore.transfer(wallet, receiver, amount);
        BigInteger transfered;
        if (success) {
            assertSuccess(txHash);
            transfered = amount;
        } else {
            assertFailure(txHash);
            transfered = BigInteger.ZERO;
        }
        assertEquals(ecoBal.subtract(transfered), txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS));
        assertEquals(receiverBal.add(transfered), txHandler.getBalance(receiver));
    }

    @Test
    void setAdmin() throws Exception {
        BigInteger TRANSFER_AMOUNT = BigInteger.ONE;
        var adminWallet = wallets[2];
        Address receiver = KeyWallet.create().getAddress();
        _transferAndCheck(ecoOwner, receiver, TRANSFER_AMOUNT, true);
        _transferAndCheck(adminWallet, receiver, TRANSFER_AMOUNT, false);

        var txHash = ecoScore.setAdmin(adminWallet, ecoOwner.getAddress());
        assertFailure(txHash);

        txHash = ecoScore.setAdmin(ecoOwner, adminWallet.getAddress());
        assertSuccess(txHash);

        _transferAndCheck(ecoOwner, receiver, TRANSFER_AMOUNT, false);
        _transferAndCheck(adminWallet, receiver, TRANSFER_AMOUNT, true);

        txHash = ecoScore.setAdmin(ecoOwner, adminWallet.getAddress());
        assertFailure(txHash);

        txHash = ecoScore.setAdmin(adminWallet, ecoOwner.getAddress());
        assertSuccess(txHash);

        _transferAndCheck(ecoOwner, receiver, TRANSFER_AMOUNT, true);
        _transferAndCheck(adminWallet, receiver, TRANSFER_AMOUNT, false);
    }
}