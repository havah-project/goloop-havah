package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Bytes;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.ChainScore;
import io.havah.test.score.SustainableFundScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Order;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;

@Tag(Constants.TAG_HAVAH)
public class SustainableFundTest extends TestBase {
    private static IconService iconService;
    private static TransactionHandler txHandler;
    private static Wallet[] wallets;
    private static SustainableFundScore sfScore;
    private static ChainScore chainScore;
    private static Wallet sfOwner;
    private static BigInteger totalRewardPerDay
            = BigInteger.valueOf(430).multiply(BigInteger.TEN.pow(4)).multiply(BigInteger.TEN.pow(18));

    @BeforeAll
    static void setup() throws Exception {
        iconService = Utils.getIconService();
        txHandler = Utils.getTxHandler();
        wallets = new Wallet[10];

        // init wallets
        sfOwner = Utils.getGovernor();
        wallets[0] = sfOwner;
        for (int i = 1; i < wallets.length; i++) {
            wallets[i] = KeyWallet.create();
        }
        Utils.distributeCoin(wallets);
        sfScore = new SustainableFundScore(txHandler);
        chainScore = new ChainScore(txHandler);

        var startHeight = Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(startHeight);

        startHeight = Utils.getHeight().add(BigInteger.valueOf(5));
        var govScore = Utils.getGovScore();
        var governor = Utils.getGovernor();
        govScore.startRewardIssue(governor, startHeight);
        Utils.waitUtil(startHeight.add(BigInteger.ONE));

        startHeight = Utils.getHeight().add(BigInteger.valueOf(5));
        govScore.startRewardIssue(governor, startHeight);
        Utils.waitUtil(startHeight);
    }

    /*

     */
    @Test
    @Order(1)
    void checkTxFee() throws Exception {
        // inflow - check tx_fee
        // inflow - failed_reward
        var treasuryBalance = txHandler.getBalance(Constants.TREASURY_ADDRESS);
        var sfBalance = txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
        var ecoBalance = txHandler.getBalance(Constants.ECOSYSTEM_ADDRESS);

        BigInteger amount = BigInteger.TEN;
        var walletNum = wallets.length;
        Bytes[] txHash = new Bytes[walletNum];
        Wallet tmp = KeyWallet.create();
        for (int i = 0; i < walletNum; i++) {
            txHash[i] = txHandler.transfer(wallets[i], tmp.getAddress(), amount);
        }
        var tmpBal = BigInteger.ZERO;
        var txFee = BigInteger.ZERO;
        for (int i = 0; i < walletNum; i++) {
            var result = txHandler.getResult(txHash[i]);
            txFee = txFee.add(result.getStepPrice().multiply(result.getStepUsed()));
            assertEquals(BigInteger.ONE, result.getStatus());
            tmpBal = tmpBal.add(amount);
        }
        assertEquals(tmpBal, txHandler.getBalance(tmp.getAddress()));

        var cmptreasuryBalance = txHandler.getBalance(Constants.TREASURY_ADDRESS);
        assertEquals(treasuryBalance.add(txFee), cmptreasuryBalance);
        treasuryBalance = cmptreasuryBalance;
        var height = Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(height);
        var inflow = sfScore.getInflow();
        Utils.waitUtilNextTerm();
        Utils.waitUtilNextTerm();
        var inflow2 = sfScore.getInflow();
        Map<SF_INFLOW, BigInteger> addedAmount = Map.of(
                SF_INFLOW.TX_FEE, treasuryBalance.multiply(BigInteger.valueOf(80)).divide(BigInteger.valueOf(100)),
                SF_INFLOW.MISSING_REWARD, totalRewardPerDay,
                SF_INFLOW.SERVICE_FEE, BigInteger.ZERO,
                SF_INFLOW.PLANET_SALES, BigInteger.ZERO
        );
        for (var type : SF_INFLOW.values()) {
            var value = inflow.getItem(type.getTypeName()).asInteger();
            var value2 = inflow2.getItem(type.getTypeName()).asInteger();
            assertEquals(value.add(addedAmount.get(type)), value2);
        }
        // outflow - hoover_refill
    }

    @Test
    void transferToken() {
        // must deploy erc20
        // transfer to SF
        // transferToken
    }

    private void _transferAndCheck(Wallet wallet, boolean expectedResult) throws Exception {
        boolean success = wallet.equals(sfOwner);
        assertEquals(expectedResult, success);

       var sfBalance = iconService.getBalance(Constants.SUSTAINABLEFUND_ADDRESS).execute();
        final var amount = BigInteger.valueOf(500);
        var addr = wallets[2].getAddress();
        var walletBalance = iconService.getBalance(addr).execute();
        var result = txHandler.getResult(sfScore.transfer(wallet, addr, amount));
        assertEquals(success ? BigInteger.ONE : BigInteger.ZERO, result.getStatus());
        walletBalance = success ? walletBalance.add(amount) : walletBalance;
        assertEquals(walletBalance, iconService.getBalance(addr).execute());
        sfBalance = success ? sfBalance.subtract(amount) : sfBalance;
        assertEquals(sfBalance, iconService.getBalance(Constants.SUSTAINABLEFUND_ADDRESS).execute());
    }

    @Test
    void transfer() throws Exception {
        var balance = iconService.getBalance(Constants.SUSTAINABLEFUND_ADDRESS).execute();
        final var amount = BigInteger.valueOf(500);
        var result = txHandler.getResult(txHandler.transfer(wallets[0], Constants.SUSTAINABLEFUND_ADDRESS, amount));
        assertEquals(BigInteger.ONE, result.getStatus());
        balance = balance.add(amount);
        assertEquals(balance, iconService.getBalance(Constants.SUSTAINABLEFUND_ADDRESS).execute());

        _transferAndCheck(wallets[1], false);
        _transferAndCheck(sfOwner, true);
    }

    // 0 - planet sales
    // 1 - service fee
    Bytes _inflowFrom(int type, Wallet wallet, BigInteger value) throws Exception {
        Bytes txHash;
        if (type == 0) {
            txHash = sfScore.transferFromPlanetSales(wallet, value);
        } else {
            txHash = sfScore.transferFromServiceFee(wallet, value);
        }
        return txHash;
    }

    @Test
    void checkInflow() throws Exception {
        var inflowObj = sfScore.getInflow();
        List<Bytes> txHashList = new ArrayList<>();
        var bal = txHandler.getBalance(wallets[2].getAddress());
        txHashList.add(_inflowFrom(0, wallets[2], BigInteger.ONE));
        bal.add(BigInteger.ONE);
        txHashList.add(_inflowFrom(1, wallets[2], BigInteger.TWO));
        bal.add(BigInteger.TWO);
        BigInteger txFee = BigInteger.ZERO;
        for (var tx : txHashList) {
            var result = txHandler.getResult(tx);
            assertEquals(BigInteger.ONE, result.getStatus(), result.toString());
            txFee = txFee.add(result.getStepUsed().multiply(result.getStepPrice()));
        }
        assertEquals(bal.subtract(BigInteger.valueOf(3)).subtract(txFee), txHandler.getBalance(wallets[2].getAddress()));
        var inflowObj2 = sfScore.getInflow();
        assertEquals(inflowObj.getItem(SF_INFLOW.PLANET_SALES.getTypeName()).asInteger().add(BigInteger.ONE),
                inflowObj2.getItem(SF_INFLOW.PLANET_SALES.getTypeName()).asInteger());
        assertEquals(inflowObj.getItem(SF_INFLOW.SERVICE_FEE.getTypeName()).asInteger().add(BigInteger.TWO),
                inflowObj2.getItem(SF_INFLOW.SERVICE_FEE.getTypeName()).asInteger());
    }

    enum SF_INFLOW {
        TX_FEE("TX_FEE"),
        SERVICE_FEE("SERVICE_FEE"),
        MISSING_REWARD("MISSING_REWARD"),
        PLANET_SALES("PLANET_SALES");

        private final String typeName;

        SF_INFLOW(String type) {
            this.typeName = type;
        }

        public String getTypeName() {
            return this.typeName;
        }
    }
}
