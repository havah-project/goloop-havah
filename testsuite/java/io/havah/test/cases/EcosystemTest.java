package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
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
public class EcosystemTest {
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
        assertEquals(BigInteger.ONE, result.getStatus());
    }

    // search concurrent in junit
    @Test
    public void checkLockupSchedule() throws Exception {
        var schedule = ecoScore.getLockupSchedule();
        List<LockSchedule> list = new ArrayList<>();
        for (var s : schedule) {
            LOG.info("s(" + s + ")");
            var height = s.asObject().getItem("BLOCK_HEIGHT").asInteger();
            var amount = s.asObject().getItem("LOCKUP_AMOUNT").asInteger();
            list.add(new LockSchedule(height, amount));
        }
        list.sort(Comparator.comparing(a -> a.blockHeight));
        var lockedAmount = Constants.ECOSYSTEM_INITIAL_BALANCE;
        Address receiver = wallets[2].getAddress();
        for (var l : list) {
            var cur = Utils.getHeight();
            if (cur.compareTo(l.blockHeight) < 0) {
                LOG.info("curHeight(" + cur + ")");
                // failure case - not owner wallet
                final BigInteger testAmount = BigInteger.valueOf(1);
                var result = txHandler.getResult(ecoScore.transfer(wallets[1], receiver, testAmount));
                assertEquals(BigInteger.ZERO, result.getStatus());

                // success case - owner wallet
                var ecoBalance = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);
                result = txHandler.getResult(ecoScore.transfer(ecoOwner, receiver, testAmount));
                boolean available = ecoBalance.subtract(lockedAmount).compareTo(testAmount) >= 0;
                assertEquals(available ? BigInteger.ONE : BigInteger.ZERO, result.getStatus(),
                        "balance(" + ecoBalance + "), locked(" + lockedAmount + "), available(" + available + ")");

                 _transferExceedAndMax(receiver, lockedAmount);

                Utils.waitUtil(l.blockHeight);
                LOG.info("lock amount(" + l.amount + ") from (" + l.blockHeight + ") height");
                lockedAmount = l.amount;
            } else {
                lockedAmount = BigInteger.ZERO;
            }
        }
        _transferExceedAndMax(receiver, lockedAmount);
    }

    @Test
    void transfer() throws Exception {

    }

    static class LockSchedule {
        private final BigInteger blockHeight;
        private final BigInteger amount;
        public LockSchedule(BigInteger blockHeight, BigInteger amount) {
            this.blockHeight = blockHeight;
            this.amount = amount;
        }

        @Override
        public String toString() {
            return "LockSchedule{" +
                    "blockHeight=" + blockHeight +
                    ", amount=" + amount +
                    '}';
        }
    }
}
