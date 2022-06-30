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
    private static IconService iconService;
    private static TransactionHandler txHandler;
    private static EcosystemScore ecoScore;
    private static Wallet ecoOwner;
    private static Wallet[] wallets;

    @BeforeAll
    static void setup() throws Exception {
        iconService = Utils.getIconService();
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

    private void _transferExceedAndMax(Address receiver, BigInteger amount) throws Exception {
        // test transfer exceed amount
        LOG.info("_transferExceedAndMax balance(" + txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS) + "), amount(" + amount + ")");
        var result = txHandler.getResult(ecoScore.transfer(ecoOwner, receiver, amount.add(BigInteger.ONE)));
        assertEquals(BigInteger.ZERO, result.getStatus());

        // test transfer all transferable
        result = txHandler.getResult(ecoScore.transfer(ecoOwner, receiver, amount));
        assertEquals(BigInteger.ONE, result.getStatus());
    }

    @Test
    public void schedule() throws Exception {
        var schedule = ecoScore.getLockupSchedule();
        List<LockSchedule> list = new ArrayList<>();
        for (var s : schedule) {
            LOG.info("s(" + s + ")");
            var height = s.asObject().getItem("BLOCK_HEIGHT").asInteger();
            var amount = s.asObject().getItem("LOCKUP_AMOUNT").asInteger();
            list.add(new LockSchedule(height, amount));
        }
        list.sort(Comparator.comparing(a -> a.blockHeight));
        var transferable = BigInteger.ZERO;
        Address receiver = wallets[2].getAddress();
        for (var l : list) {
            var cur = Utils.getHeight();
            if (cur.compareTo(l.blockHeight) < 0) {
                LOG.info("curHeight(" + cur + ")");
                // failure case
                final BigInteger testAmount = BigInteger.valueOf(5);
                var result = txHandler.getResult(ecoScore.transfer(wallets[1], receiver, testAmount));
                assertEquals(BigInteger.ZERO, result.getStatus());
                result = txHandler.getResult(ecoScore.transfer(ecoOwner, receiver, testAmount));
                boolean available = transferable.compareTo(testAmount) >= 0;
                assertEquals(available ? BigInteger.ONE : BigInteger.ZERO, result.getStatus());
                if (available) {
                    transferable = transferable.subtract(testAmount);
                }

                if (transferable.compareTo(BigInteger.ZERO) > 0) {
                    _transferExceedAndMax(receiver, transferable);
                    transferable = transferable.subtract(transferable);
                }

                Utils.waitUtil(l.blockHeight);
                transferable = transferable.add(l.amount);
            }
        }
        _transferExceedAndMax(receiver, transferable);
        assertEquals(BigInteger.ZERO, txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS));
    }

    static class LockSchedule {
        private BigInteger blockHeight;
        private BigInteger amount;
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
