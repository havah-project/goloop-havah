package io.havah.test.cases;

import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
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
import java.util.Map;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;

@Tag(Constants.TAG_HAVAH)
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
public class VaultTest extends TestBase {
    private static TransactionHandler txHandler;
    private static KeyWallet[] wallets;
    private static VaultScore vaultScore;

    private static Wallet ownerWallet;

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

        ownerWallet = Utils.getGovernor();
    }

    void _checkAndClaim(KeyWallet wallet) throws IOException, ResultTimeoutException {
        LOG.infoEntering("_checkAndClaim", "claim : " + wallet.getAddress());
        BigInteger claimable = vaultScore.getClaimable(wallet.getAddress());
        BigInteger balance = txHandler.getBalance(wallet.getAddress());
        TransactionResult result = vaultScore.claim(wallet);
        assertSuccess(result);
        BigInteger fee = Utils.getTxFee(result);
        balance = txHandler.getBalance(wallet.getAddress()).subtract(balance);
        LOG.info("claimable : " + claimable);
        claimable = claimable.subtract(fee);
        LOG.info("real claim amount (exclude txFee) : " + claimable);
        LOG.info("balance : " + balance);
        assertEquals(0, balance.compareTo(claimable), "claimable is not expected");
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
        vaultScore.addAllocation(ownerWallet, accounts);
        LOG.infoExiting();

        LOG.infoEntering("get", "getAllocation()");
        assertEquals(allocations[0], vaultScore.getAllocation(wallets[0].getAddress()));
        assertEquals(allocations[1], vaultScore.getAllocation(wallets[1].getAddress()));
        assertEquals(BigInteger.ZERO, vaultScore.getAllocation(wallets[2].getAddress()));
        LOG.infoExiting();

        assertEquals(BigInteger.ZERO, vaultScore.getClaimable(wallets[0].getAddress()));

        LOG.infoEntering("call", "setVestingSchedules()");
        BigInteger curTimestamp = Utils.getTimestamp();
        BigInteger[] timeStamps = { curTimestamp.add(BigInteger.valueOf(1500)), curTimestamp.add(BigInteger.valueOf(3000)), curTimestamp.add(BigInteger.valueOf(4500)) };
        VaultScore.VestingSchedule[] schedules = {
                new VaultScore.VestingSchedule(timeStamps[0], BigInteger.valueOf(100), BigInteger.valueOf(25)),
                new VaultScore.VestingSchedule(timeStamps[1], BigInteger.valueOf(100), BigInteger.valueOf(50)),
                new VaultScore.VestingSchedule(timeStamps[2], BigInteger.valueOf(100), BigInteger.valueOf(100))
        };
        vaultScore.setVestingSchedules(ownerWallet, wallets[0].getAddress(), schedules);
        vaultScore.setVestingSchedules(ownerWallet, wallets[1].getAddress(), schedules);
        LOG.infoExiting();

        LOG.infoEntering("claim", "claim vault");
        for(int i=0; i<schedules.length; i++) {
            Utils.waitUtilTime(timeStamps[i]);
            _checkAndClaim(wallets[0]);
        }
        _checkAndClaim(wallets[1]);

        TransactionResult result = vaultScore.claim(wallets[2]);
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
        var txHash = vaultScore.setAdmin(wallets[1], ownerWallet.getAddress());
        assertFailure(txHash);

        txHash = vaultScore.setAdmin(ownerWallet, wallets[1].getAddress());
        assertSuccess(txHash);
        LOG.infoExiting();

        LOG.infoEntering("call", "admin()");
        assertEquals(false, vaultScore.admin().equals(ownerWallet.getAddress()));
        assertEquals(true, vaultScore.admin().equals(wallets[1].getAddress()));
        LOG.infoExiting();

        LOG.infoEntering("call", "addAllocation()");
        VaultScore.VestingAccount[] accounts = {
                new VaultScore.VestingAccount(tmpWallets[0].getAddress(), BigInteger.ZERO),
                new VaultScore.VestingAccount(tmpWallets[1].getAddress(), BigInteger.ZERO),
                new VaultScore.VestingAccount(tmpWallets[2].getAddress(), BigInteger.ZERO)
        };
        assertFailure(vaultScore.addAllocation(ownerWallet, accounts));
        assertSuccess(vaultScore.addAllocation(wallets[1], accounts));
        LOG.infoExiting();

        LOG.infoEntering("call", "setAllocation()");
        assertFailure(vaultScore.setAllocation(ownerWallet, new VaultScore.VestingAccount(wallets[0].getAddress(), BigInteger.ZERO)));
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
        assertFailure(vaultScore.setVestingSchedules(ownerWallet, wallets[0].getAddress(), schedules));
        assertSuccess(vaultScore.setVestingSchedules(wallets[1], wallets[1].getAddress(), schedules));
        LOG.infoExiting();

        txHash = vaultScore.setAdmin(wallets[1], ownerWallet.getAddress());
        assertSuccess(txHash);
    }
}
