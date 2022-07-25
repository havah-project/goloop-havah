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
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.math.BigInteger;
import java.util.Map;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;

@Tag(Constants.TAG_HAVAH)
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

    BigInteger _getVaultClaimableAmount(Address address) throws IOException {
        var vestingInfo = vaultScore.getVestingInfo(address);
        if(vestingInfo.containsKey("claimable")) {
            return (BigInteger) vestingInfo.get("claimable");
        }
        return BigInteger.ZERO;
    }

    Map<String, Object> _getVestingInfo(Address address) throws IOException {
        return vaultScore.getVestingInfo(address);
    }

    void _checkAndClaim(KeyWallet wallet) throws IOException, ResultTimeoutException {
        LOG.infoEntering("_checkAndClaim", "claim : " + wallet.getAddress());
        BigInteger claimable = _getVaultClaimableAmount(wallet.getAddress());
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
    void startVault() throws Exception {
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));

        LOG.infoEntering("call", "addAllocation()");
        VaultScore.VestingAccount[] accounts = {
                new VaultScore.VestingAccount(wallets[0].getAddress(), ICX.multiply(BigInteger.valueOf(150))),
                new VaultScore.VestingAccount(wallets[1].getAddress(), ICX.multiply(BigInteger.valueOf(100))),
                new VaultScore.VestingAccount(wallets[2].getAddress(), ICX.multiply(BigInteger.valueOf(50)))
        };
        vaultScore.addAllocation(ownerWallet, accounts);
        LOG.infoExiting();

        LOG.infoEntering("call", "setVestingHeights()");
        BigInteger curHeight = Utils.getHeight();
        BigInteger[] heights = { curHeight.add(BigInteger.valueOf(10)), curHeight.add(BigInteger.valueOf(20)), curHeight.add(BigInteger.valueOf(30)) };
        VaultScore.VestingHeight[] schedules = {
                new VaultScore.VestingHeight(heights[0], BigInteger.valueOf(2500)),
                new VaultScore.VestingHeight(heights[1], BigInteger.valueOf(5000)),
                new VaultScore.VestingHeight(heights[2], BigInteger.valueOf(10000))
        };
        vaultScore.setVestingHeights(ownerWallet, schedules);
        LOG.infoExiting();

        LOG.infoEntering("transfer icx", "transfer 100 ICX to vault score from account");
        Bytes txHash = txHandler.transfer(wallets[0], vaultScore.getAddress(), ICX.multiply(BigInteger.valueOf(100)));
        assertFailure(txHandler.getResult(txHash));
        LOG.infoExiting();

        LOG.info("vestingInfo wallets[0] : " + _getVestingInfo(wallets[0].getAddress()));

        LOG.infoEntering("claim", "claim vault");
        for(int i=0; i<heights.length; i++) {
            Utils.waitUtil(heights[i]);
            _checkAndClaim(wallets[0]);
        }
        LOG.info("vestingInfo wallets[1] : " + _getVestingInfo(wallets[1].getAddress()));
        _checkAndClaim(wallets[1]);
        LOG.infoExiting();
    }
}
