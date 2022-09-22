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
import static org.junit.jupiter.api.Assertions.assertEquals;

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
        wallets = new KeyWallet[3];
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
        Bytes txHash = txHandler.transfer(wallets[0], vaultScore.getAddress(), ICX.multiply(BigInteger.valueOf(100)));
        assertFailure(txHandler.getResult(txHash));
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
        StringBuilder builder = new StringBuilder();
        builder.append("[");
        for (int i=0; i<accounts.length; i++) {
            if(i > 0)
                builder.append(',');
            builder.append(accounts[i].toString());
        }
        builder.append("]");
        vaultScore.ensureAddAllocation(result, builder.toString());
        LOG.infoExiting();

        LOG.infoEntering("get", "getAllocation()");
        assertEquals(allocations[0], vaultScore.getAllocation(wallets[0].getAddress()));
        assertEquals(allocations[1], vaultScore.getAllocation(wallets[1].getAddress()));
        assertEquals(null, vaultScore.getAllocation(wallets[2].getAddress()));
        LOG.infoExiting();

        assertEquals(BigInteger.ZERO, vaultScore.getClaimable(wallets[0].getAddress()));

        LOG.infoEntering("call", "setAllocation()");
        VaultScore.VestingAccount vestingAccount = new VaultScore.VestingAccount(wallets[0].getAddress(), allocations[0]);
        assertFailure(vaultScore.setAllocation(wallets[0], vestingAccount));
        result = vaultScore.setAllocation(governorWallet, vestingAccount);
        vaultScore.ensureSetAllocation(result, vestingAccount.toString());
        LOG.infoExiting();

        LOG.infoEntering("call", "setVestingSchedules()");
        BigInteger curTimestamp = Utils.getTimestamp();
        BigInteger[] timeStamps = {
                curTimestamp.add(BigInteger.valueOf(20000000)),
                curTimestamp.add(BigInteger.valueOf(40000000)),
                curTimestamp.add(BigInteger.valueOf(60000000))
        };
        VaultScore.VestingSchedule[] schedules = {
                new VaultScore.VestingSchedule(timeStamps[0], BigInteger.valueOf(100), BigInteger.valueOf(25)),
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(50)),
                new VaultScore.VestingSchedule(timeStamps[2], BigInteger.valueOf(100), BigInteger.valueOf(100))
        };
        VaultScore.VestingSchedule[] falseSchedules = {
                new VaultScore.VestingSchedule(timeStamps[0], BigInteger.valueOf(25), BigInteger.valueOf(15)),
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(15)),
                new VaultScore.VestingSchedule(timeStamps[2], BigInteger.valueOf(100), BigInteger.valueOf(100))
        };
        VaultScore.VestingSchedule[] falseSchedules2 = {
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(150))
        };

        assertFailure(vaultScore.setVestingSchedules(wallets[0], wallets[0].getAddress(), schedules));
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[2].getAddress(), schedules));
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), falseSchedules));
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), falseSchedules2));
        result = vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), schedules);
        assertSuccess(result);
        builder = new StringBuilder();
        builder.append("[");
        for (int i=0; i<schedules.length; i++) {
            if(i > 0)
                builder.append(',');
            builder.append(schedules[i].toString());
        }
        builder.append("]");
        vaultScore.ensureSetVestingSchedules(result, wallets[0].getAddress(), builder.toString());
        result = vaultScore.setVestingSchedules(governorWallet, wallets[1].getAddress(), schedules);
        assertSuccess(result);
        vaultScore.ensureSetVestingSchedules(result, wallets[1].getAddress(), builder.toString());
        LOG.infoExiting();

        LOG.infoEntering("claim", "claim vault");
        for(int i=0; i<schedules.length; i++) {
            Utils.waitUtilTime(schedules[i].timestamp);
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
        KeyWallet[] tmpWallets = new KeyWallet[3];
        for (int i = 0; i < tmpWallets.length; i++) {
            tmpWallets[i] = KeyWallet.create();
        }

        LOG.infoEntering("call", "setAdmin()");
        var txHash = vaultScore.setAdmin(wallets[1], governorWallet.getAddress());
        assertFailure(txHash);

        txHash = vaultScore.setAdmin(governorWallet, wallets[1].getAddress());
        assertSuccess(txHash);
        LOG.infoExiting();

        LOG.infoEntering("call", "admin()");
        assertEquals(false, vaultScore.admin().equals(governorWallet.getAddress()));
        assertEquals(true, vaultScore.admin().equals(wallets[1].getAddress()));
        LOG.infoExiting();

        LOG.infoEntering("call", "addAllocation()");
        VaultScore.VestingAccount[] accounts = {
                new VaultScore.VestingAccount(tmpWallets[0].getAddress(), BigInteger.ZERO),
                new VaultScore.VestingAccount(tmpWallets[1].getAddress(), BigInteger.ZERO),
                new VaultScore.VestingAccount(tmpWallets[2].getAddress(), BigInteger.ZERO)
        };
        assertFailure(vaultScore.addAllocation(governorWallet, accounts));
        assertSuccess(vaultScore.addAllocation(wallets[1], accounts));
        LOG.infoExiting();

        LOG.infoEntering("call", "setAllocation()");
        assertFailure(vaultScore.setAllocation(governorWallet, new VaultScore.VestingAccount(wallets[0].getAddress(), BigInteger.ZERO)));
        assertSuccess(vaultScore.setAllocation(wallets[1], new VaultScore.VestingAccount(wallets[0].getAddress(), BigInteger.ZERO)));
        LOG.infoExiting();

        LOG.infoEntering("call", "setVestingSchedules()");
        BigInteger curTimestamp = Utils.getTimestamp();
        BigInteger[] timeStamps = { curTimestamp.add(BigInteger.valueOf(1500)), curTimestamp.add(BigInteger.valueOf(3000)), curTimestamp.add(BigInteger.valueOf(4500)) };
        VaultScore.VestingSchedule[] schedules = {
                new VaultScore.VestingSchedule(timeStamps[0], BigInteger.valueOf(100), BigInteger.valueOf(25)),
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(50)),
                new VaultScore.VestingSchedule(timeStamps[2], BigInteger.valueOf(100), BigInteger.valueOf(100))
        };
        assertFailure(vaultScore.setVestingSchedules(governorWallet, wallets[0].getAddress(), schedules));
        assertSuccess(vaultScore.setVestingSchedules(wallets[1], wallets[1].getAddress(), schedules));
        LOG.infoExiting();

        txHash = vaultScore.setAdmin(wallets[1], governorWallet.getAddress());
        assertSuccess(txHash);
    }
}
