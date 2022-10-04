package io.havah.test.cases;

import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.test.common.ResultTimeoutException;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.VaultScore;
import org.junit.jupiter.api.*;

import java.io.IOException;
import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.*;

@Tag(Constants.TAG_HAVAH)
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
public class VaultTest extends TestBase {
    private static TransactionHandler txHandler;
    private static KeyWallet[] wallets;
    private static VaultScore vaultScore;

    private static Wallet governorWallet;

    @BeforeAll
    static void setup() throws Exception {
        txHandler = Utils.getTxHandler();
        wallets = new KeyWallet[4];
        BigInteger amount = ICX.multiply(BigInteger.valueOf(300));
        for (int i = 0; i < wallets.length; i++) {
            wallets[i] = KeyWallet.create();
            txHandler.transfer(wallets[i].getAddress(), amount);
        }
        for (KeyWallet wallet : wallets) {
            ensureIcxBalance(txHandler, wallet.getAddress(), BigInteger.ZERO, amount);
        }

        vaultScore = new VaultScore(txHandler);

        governorWallet = Utils.getGovernor();
    }

    void _checkAndClaim(KeyWallet wallet) throws IOException, ResultTimeoutException {
        LOG.infoEntering("_checkAndClaim", "claim : " + wallet.getAddress());
        BigInteger before = txHandler.getBalance(wallet.getAddress());
        LOG.info("balance before : " + before);
        BigInteger claimable = vaultScore.getClaimable(wallet.getAddress());
        TransactionResult result = vaultScore.claim(wallet);
        assertSuccess(result);
        vaultScore.ensureClaim(result, wallet.getAddress(), claimable);
        BigInteger fee = Utils.getTxFee(result);
        BigInteger after = txHandler.getBalance(wallet.getAddress());
        LOG.info("claimable : " + claimable);
        claimable = claimable.subtract(fee);
        LOG.info("claimed amount (exclude txFee) : " + claimable);
        LOG.info("balance after : " + after);
        assertEquals(0, after.subtract(before).compareTo(claimable), "claimable is not expected");
        LOG.infoExiting();
    }

    @Test
    @Order(1)
    void startVault() throws Exception {
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));

        LOG.infoEntering("transfer icx", "transfer 100 ICX to vault score from account");
        Bytes txHash = txHandler.transfer(wallets[0], vaultScore.getAddress(), ICX.multiply(BigInteger.valueOf(250)));
        assertSuccess(txHandler.getResult(txHash));
        LOG.infoExiting();

        LOG.infoEntering("call", "addAllocation()");
        BigInteger[] allocations = { ICX.multiply(BigInteger.valueOf(150)), ICX.multiply(BigInteger.valueOf(100))};
        VaultScore.VestingAccount[] accounts = {
                new VaultScore.VestingAccount(wallets[0].getAddress(), allocations[0]),
                new VaultScore.VestingAccount(wallets[1].getAddress(), allocations[1])
        };
        assertFailure(vaultScore.addAllocation(wallets[0], accounts));
        assertFailure(vaultScore.addAllocation(governorWallet, new VaultScore.VestingAccount[] {}));
        TransactionResult result = vaultScore.addAllocation(governorWallet, accounts);
        assertSuccess(result);
        LOG.infoExiting();

        LOG.infoEntering("get", "getAllocation()");
        assertEquals(allocations[0], vaultScore.getAllocation(wallets[0].getAddress()));
        assertEquals(allocations[1], vaultScore.getAllocation(wallets[1].getAddress()));
        boolean exception = false;
        try {
            var val = vaultScore.getAllocation(wallets[2].getAddress());
        } catch (Exception ex){
            LOG.info("val(" + ex + ")");
            exception = true;
        }
        assertTrue(exception);
//        assertEquals(null, vaultScore.getAllocation(wallets[2].getAddress()));
        LOG.infoExiting();

        assertEquals(BigInteger.ZERO, vaultScore.getClaimable(wallets[0].getAddress()));

        LOG.infoEntering("call", "setAllocation()");
        VaultScore.VestingAccount vestingAccount = new VaultScore.VestingAccount(wallets[0].getAddress(), allocations[0]);
        assertFailure(vaultScore.setAllocation(wallets[0], vestingAccount));
        result = vaultScore.setAllocation(governorWallet, vestingAccount);
        assertSuccess(result);
        LOG.infoExiting();

        LOG.infoEntering("call", "setVestingSchedules()");
        BigInteger curTimestamp = Utils.getTimestamp();
        BigInteger[] timeStamps = {
                curTimestamp.add(BigInteger.valueOf(40000000)),
                curTimestamp.add(BigInteger.valueOf(60000000)),
                curTimestamp.add(BigInteger.valueOf(80000000))
        };
        VaultScore.VestingSchedule[] successSchedules1 = {
                new VaultScore.VestingSchedule(timeStamps[0], BigInteger.valueOf(100), BigInteger.valueOf(25)),
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(50)),
                new VaultScore.VestingSchedule(timeStamps[2], BigInteger.valueOf(100), BigInteger.valueOf(100))
        };
        VaultScore.VestingSchedule[] successSchedules2 = {
                new VaultScore.VestingSchedule(timeStamps[0], BigInteger.valueOf(100), BigInteger.valueOf(0)),
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(50))
        };
        VaultScore.VestingSchedule[] failureSchedules1 = {
                new VaultScore.VestingSchedule(timeStamps[0], BigInteger.valueOf(25), BigInteger.valueOf(15)),
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(15))
        };
        VaultScore.VestingSchedule[] failureSchedules2 = {
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(150))
        };
        VaultScore.VestingSchedule[] failureSchedules3 = {
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(0), BigInteger.valueOf(150))
        };
        VaultScore.VestingSchedule[] failureSchedules4 = {
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(-1))
        };

        assertFailure(vaultScore.setVestingSchedules(wallets[0], wallets[0].getAddress(), successSchedules1));
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[2].getAddress(), successSchedules1));
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), failureSchedules1));
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), failureSchedules2));
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), failureSchedules3));
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), failureSchedules4));
        result = vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), successSchedules2);
        assertSuccess(result);
        var s =  vaultScore.getSchedule(wallets[0].getAddress());
        LOG.info("schedule(" + s + ")");
        assertEquals(successSchedules2.length, s.size());
        result = vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), successSchedules1);
        assertSuccess(result);
        s =  vaultScore.getSchedule(wallets[0].getAddress());
        int size = s.size();
        LOG.info("schedule(" + s + ")");
        assertEquals(successSchedules1.length, s.size());
        result = vaultScore.setVestingSchedules(governorWallet, wallets[1].getAddress(), successSchedules2);
        assertSuccess(result);
        LOG.infoExiting();

        LOG.infoEntering("claim", "claim vault");
        for(int i=0; i<successSchedules1.length; i++) {
            Utils.waitUtilTime(successSchedules1[i].timestamp);
            _checkAndClaim(wallets[0]);
        }
        _checkAndClaim(wallets[1]);

        result = vaultScore.claim(wallets[2]);
        assertFailure(result);

        LOG.infoExiting();
    }

    @Test
    @Order(2)
    void setAdmin() throws Exception {
        LOG.infoEntering("call", "setAdmin()");
        var newAdmin =  wallets[1];
        var txHash = vaultScore.setAdmin(newAdmin, governorWallet.getAddress());
        assertFailure(txHash);

        txHash = vaultScore.setAdmin(governorWallet, wallets[1].getAddress());
        assertSuccess(txHash);
        LOG.infoExiting();

        LOG.infoEntering("call", "admin()");
        assertNotEquals(vaultScore.admin(), governorWallet.getAddress());
        assertEquals(vaultScore.admin(), newAdmin.getAddress());
        LOG.infoExiting();

        LOG.infoEntering("call", "addAllocation()");
        VaultScore.VestingAccount[] accounts = {
                new VaultScore.VestingAccount(wallets[2].getAddress(), BigInteger.ZERO),
                new VaultScore.VestingAccount(wallets[3].getAddress(), BigInteger.ZERO)
        };
        assertFailure(vaultScore.addAllocation(governorWallet, accounts));
        assertSuccess(vaultScore.addAllocation(newAdmin, accounts));
        LOG.infoExiting();

        LOG.infoEntering("call", "setAllocation()");
        assertFailure(vaultScore.setAllocation(governorWallet, new VaultScore.VestingAccount(wallets[2].getAddress(), BigInteger.ZERO)));
        assertSuccess(vaultScore.setAllocation(newAdmin, new VaultScore.VestingAccount(wallets[2].getAddress(), BigInteger.ZERO)));
        LOG.infoExiting();

        LOG.infoEntering("call", "setVestingSchedules()");
        BigInteger curTimestamp = Utils.getTimestamp();
        BigInteger[] timeStamps = { curTimestamp.add(BigInteger.valueOf(1500)), curTimestamp.add(BigInteger.valueOf(3000)), curTimestamp.add(BigInteger.valueOf(4500)) };
        VaultScore.VestingSchedule[] schedules = {
                new VaultScore.VestingSchedule(timeStamps[0], BigInteger.valueOf(100), BigInteger.valueOf(25)),
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(50)),
                new VaultScore.VestingSchedule(timeStamps[2], BigInteger.valueOf(100), BigInteger.valueOf(100))
        };
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[3].getAddress(), schedules));
        assertSuccess(vaultScore.setVestingSchedules(newAdmin, wallets[3].getAddress(), schedules));
        LOG.infoExiting();

        txHash = vaultScore.setAdmin(newAdmin, governorWallet.getAddress());
        assertSuccess(txHash);
    }
}
