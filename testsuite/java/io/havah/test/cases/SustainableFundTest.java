package io.havah.test.cases;

import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.IRC2TokenScore;
import io.havah.test.score.PlanetNFTScore;
import io.havah.test.score.SustainableFundScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Order;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Random;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;

@Tag(Constants.TAG_HAVAH)
public class SustainableFundTest extends TestBase {
    private static PlanetNFTScore planetNFTScore;
    private static TransactionHandler txHandler;
    private static Wallet[] wallets;
    private static SustainableFundScore sfScore;
    private static Wallet sfOwner;

    @BeforeAll
    static void setup() throws Exception {
        txHandler = Utils.getTxHandler();
        wallets = new Wallet[10];

        // init wallets
        sfOwner = Utils.getGovernor();
        planetNFTScore = new PlanetNFTScore(Utils.getGovernor(), txHandler);
        wallets[0] = sfOwner;
        for (int i = 1; i < wallets.length; i++) {
            wallets[i] = KeyWallet.create();
        }
        Utils.distributeCoin(wallets);
        sfScore = new SustainableFundScore(txHandler);
    }

    /*

     */
    @Test
    @Order(1)
    void checkTxFee() throws Exception {
        if (Utils.isRewardIssued()) {
            LOG.info("reward issued already so skip the SustainableFund.checkTxFee");
            return;
        }
        // inflow - check tx_fee
        // inflow - failed_reward
        // 1. check treasury balance
        var treasuryBalance = txHandler.getBalance(Constants.TREASURY_ADDRESS);

        // 2. send transaction to check increased treasury balance
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

        // 3. check treasury balance -> before treasury balance + txFee
        var cmpTreasuryBalance = txHandler.getBalance(Constants.TREASURY_ADDRESS);
        assertEquals(treasuryBalance.add(txFee), cmpTreasuryBalance,
                String.format("treasury balance before(%s), after(%s), txFee(%s)", treasuryBalance, cmpTreasuryBalance, txFee));
        var sfBalance = txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
        var height = Utils.startRewardIssueIfNotStarted();
        treasuryBalance = txHandler.getBalance(Constants.TREASURY_ADDRESS);
        Utils.waitUtil(height);
        var inflow = sfScore.getInflow();
        Utils.waitUtilNextTerm();
        Utils.waitUtil(Utils.getHeightNext(1));
//        var planetAmount = planetNFTScore.totalSupply();
//        var missingReward = Constants.TOTAL_REWARD_PER_DAY.divide(planetAmount).multiply(planetAmount);
        var inflow2 = sfScore.getInflow();
        Map<SF_INFLOW, BigInteger> addedAmount = Map.of(
                SF_INFLOW.TX_FEE, treasuryBalance.multiply(BigInteger.valueOf(80)).divide(BigInteger.valueOf(100)), // 80 % of treasury
                SF_INFLOW.MISSING_REWARD, Constants.TOTAL_REWARD_PER_DAY,
                SF_INFLOW.SERVICE_FEE, BigInteger.ZERO,
                SF_INFLOW.PLANET_SALES, BigInteger.ZERO
        );
        var now = txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
        LOG.info("sfBalance before(" + sfBalance + "), after(" + now + ")");
        LOG.info("sfBalance inflow(" + inflow + "), inflow2(" + inflow2 + ")");

        for (var type : SF_INFLOW.values()) {
            var value = inflow.getItem(type.getTypeName()).asInteger();
            var value2 = inflow2.getItem(type.getTypeName()).asInteger();
            LOG.info("value(" + value + "), value2(" + value2 + "), type(" + type.getTypeName() + ")");
            assertEquals(value.add(addedAmount.get(type)), value2, String.format("type(%s), treasuryBalance(%s), cur(%s)",
                    type.getTypeName(), treasuryBalance, txHandler.getBalance(Constants.TREASURY_ADDRESS)));
        }
        // outflow - hoover_refill
    }

    @Test
    void transferToken() throws Exception {
        LOG.infoEntering("transferToken");

        // must deploy erc20
        Wallet irc2Deployer = KeyWallet.create();
        Wallet noPermission = KeyWallet.create();
        Wallet receiver = KeyWallet.create();

        LOG.info("distributeCoin");
        Utils.distributeCoin(new Wallet[]{
                irc2Deployer,
                noPermission,
                receiver,
        });

        IRC2TokenScore irc2 = IRC2TokenScore.mustDeploy(txHandler, irc2Deployer);
        LOG.info("deployed IRC2, address(" + irc2.getAddress() + ")");
        // transfer to SF
        irc2.transfer(irc2Deployer, Constants.SUSTAINABLEFUND_ADDRESS, BigInteger.TEN, null);
        // transferToken
        // failure case
        Wallet[] testWallet = new Wallet[]{
                noPermission, // failure case
                sfOwner // success case
        };
        // wrong token address
        var transferValue = BigInteger.ONE;
        byte[] addr = new byte[20];
        new Random().nextBytes(addr);
        var wrongIRC2Addr = new Address(Address.AddressPrefix.CONTRACT, addr);
        var hash = sfScore.transferToken(sfOwner, wrongIRC2Addr, receiver.getAddress(), transferValue);
        assertEquals(BigInteger.ZERO, txHandler.getResult(hash).getStatus());

        var sfTokenBalance = irc2.balanceOf(Constants.SUSTAINABLEFUND_ADDRESS);
        LOG.info("SF token balance(" + sfTokenBalance + ")");
        for (Wallet wallet : testWallet) {
            hash = sfScore.transferToken(wallet, irc2.getAddress(), receiver.getAddress(), transferValue);
            var result = txHandler.getResult(hash);
            assertEquals(wallet.equals(sfOwner) ? BigInteger.ONE : BigInteger.ZERO, result.getStatus());
            if (wallet.equals(sfOwner)) {
                // check token balance
                assertEquals(sfTokenBalance.subtract(transferValue), irc2.balanceOf(Constants.SUSTAINABLEFUND_ADDRESS));
                assertEquals(transferValue, irc2.balanceOf(receiver.getAddress()));
            }
            LOG.info((wallet.equals(sfOwner) ? "success" : "failure")
                    + " test SF token balance before : " + sfTokenBalance
                    + " after : " + irc2.balanceOf(Constants.SUSTAINABLEFUND_ADDRESS));
        }
        // success case
        LOG.infoExiting();
    }

    private void _transferAndCheck(Wallet wallet, boolean expectedResult) throws Exception {
        boolean success = wallet.equals(sfOwner);
        assertEquals(expectedResult, success);
        waitUtilNextTermIfRewardIssued();

        var sfBalance = txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
        final var amount = BigInteger.valueOf(500);
        var addr = wallets[2].getAddress();
        var walletBalance = txHandler.getBalance(addr);
        var result = txHandler.getResult(sfScore.transfer(wallet, addr, amount));
        assertEquals(success ? BigInteger.ONE : BigInteger.ZERO, result.getStatus());
        // check transferred amount from SF
        walletBalance = success ? walletBalance.add(amount) : walletBalance;
        assertEquals(walletBalance, txHandler.getBalance(addr));

        // check remained amount in SF
        var expected = success ? sfBalance.subtract(amount) : sfBalance;
        var now = txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
        assertEquals(expected, now, String.format("sf balance before(%s), after(%s)", sfBalance, now));
    }

    void waitUtilNextTermIfRewardIssued() throws Exception {
        if (Utils.isRewardIssued()) {
            Utils.waitUtilNextTerm();
            Utils.waitUtil(Utils.getHeightNext(1));
        }
    }

    @Test
    void transfer() throws Exception {
        var balance = txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
        final var amount = BigInteger.valueOf(500);
        var result = txHandler.getResult(txHandler.transfer(wallets[0], Constants.SUSTAINABLEFUND_ADDRESS, amount));
        assertEquals(BigInteger.ONE, result.getStatus());
        balance = balance.add(amount);
        assertEquals(balance, txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS));

        _transferAndCheck(wallets[1], false);
        _transferAndCheck(sfOwner, true);
    }

    // 0 - planet sales
    // 1 - service fee
    Bytes _inflowFrom(int type, Wallet wallet, BigInteger value) throws Exception {
        Bytes txHash;
        if (type == 0) {
            txHash = sfScore.depositFromPlanetSales(wallet, value);
        } else {
            txHash = sfScore.depositFromServiceFee(wallet, value);
        }
        return txHash;
    }

    private static final int DEPOSIT_FROM_PLANET_SALES = 0;
    private static final int DEPOSIT_FROM_SERVICE_FEE = 1;

    @Test
    void checkInflow() throws Exception {
        var inflowObj = sfScore.getInflow();
        var sfBalance = txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
        List<Bytes> txHashList = new ArrayList<>();
        var depositAmountFromPlanetSales = BigInteger.ONE;
        var depositAmountFromServiceFee = BigInteger.TWO;
        var depositorToSF = wallets[2];
        var bal = txHandler.getBalance(depositorToSF.getAddress());
        txHashList.add(_inflowFrom(DEPOSIT_FROM_PLANET_SALES, depositorToSF, depositAmountFromPlanetSales));
        txHashList.add(_inflowFrom(DEPOSIT_FROM_SERVICE_FEE, depositorToSF, depositAmountFromServiceFee));
        BigInteger txFee = BigInteger.ZERO;
        for (var tx : txHashList) {
            var result = txHandler.getResult(tx);
            assertEquals(BigInteger.ONE, result.getStatus(), result.toString());
            txFee = txFee.add(result.getStepUsed().multiply(result.getStepPrice()));
        }
        // check depositor balance is valid
        assertEquals(bal.subtract(
                        depositAmountFromPlanetSales.add(depositAmountFromServiceFee))
                .subtract(txFee), txHandler.getBalance(depositorToSF.getAddress()));
        var inflowObj2 = sfScore.getInflow();
        assertEquals(inflowObj.getItem(SF_INFLOW.PLANET_SALES.getTypeName()).asInteger().add(depositAmountFromPlanetSales),
                inflowObj2.getItem(SF_INFLOW.PLANET_SALES.getTypeName()).asInteger());
        assertEquals(inflowObj.getItem(SF_INFLOW.SERVICE_FEE.getTypeName()).asInteger().add(depositAmountFromServiceFee),
                inflowObj2.getItem(SF_INFLOW.SERVICE_FEE.getTypeName()).asInteger());
    }

    @Test
    void checkOutflow() throws Exception {
        if(Utils.isRewardIssued()) {
            Utils.waitUtilNextTerm();
            Utils.waitUtil(Utils.getHeightNext(1));
        }
        final String CUSTOM = SustainableFundScore.OUTFLOW_CUSTOM;
        var beforeCustom = sfScore.getOutflow().getItem(CUSTOM).asInteger();
        var sfBalance = txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
        var transferAmount = BigInteger.TEN;
        Wallet receiver = KeyWallet.create();
        var txHash = sfScore.transfer(sfOwner, receiver.getAddress(), transferAmount);
        var result = txHandler.getResult(txHash);
        assertEquals(BigInteger.ONE, result.getStatus());
        // compare balance of sustainableFund between before and after transfer
        assertEquals(sfBalance.subtract(transferAmount), txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS),
                String.format("sfBalance before(%s), after(%s)", sfBalance, txHandler.getBalance(Constants.SUSTAINABLEFUND_ADDRESS)));
        var afterCustom = sfScore.getOutflow().getItem(CUSTOM).asInteger();
        // afterCustom - beforeCustom = transferAmount
        assertEquals(transferAmount, afterCustom.subtract(beforeCustom));
        assertEquals(transferAmount, txHandler.getBalance(receiver.getAddress()));
    }

    void _setUSDT(Wallet wallet, Address address, boolean success) throws Exception {
        var result = txHandler.getResult(sfScore.setUsdt(wallet, address)); // failure
        assertEquals(success ? BigInteger.ONE : BigInteger.ZERO, result.getStatus());
    }

    void _transferTokenAndCheckResult(Wallet wallet, Address tokenAddr, Address receiver, BigInteger amount, boolean success) throws Exception {
        var txHash = sfScore.transferToken(wallet, tokenAddr, receiver, amount);
        assertEquals(success ? BigInteger.ONE : BigInteger.ZERO, txHandler.getResult(txHash).getStatus());
    }
    @Test
    void checkInflowOutflowInUSDT() throws Exception {
        // check inflowInUSDT
        RpcObject object = sfScore.getInflowInUSDT();
        var irc2Deployer = wallets[1];
        IRC2TokenScore irc2 = IRC2TokenScore.mustDeploy(txHandler, irc2Deployer);
        _setUSDT(irc2Deployer, irc2.getAddress(), false);
        _setUSDT(sfOwner, irc2.getAddress(), true);

        // transfer to SF
        final var receivedAmount = BigInteger.TEN;
        final String INFLOW_PLANETSALES = SustainableFundScore.INFLOW_PLANETSALES;
        var txHash = irc2.transfer(irc2Deployer, Constants.SUSTAINABLEFUND_ADDRESS, receivedAmount,
                INFLOW_PLANETSALES.getBytes(StandardCharsets.UTF_8));
        assertEquals(BigInteger.ONE, txHandler.getResult(txHash).getStatus());
        RpcObject afterObj = sfScore.getInflowInUSDT();
        assertEquals(object.getItem(INFLOW_PLANETSALES).asInteger().add(receivedAmount),
                afterObj.getItem(INFLOW_PLANETSALES).asInteger());

        RpcObject outflow = sfScore.getOutflowInUSDT();
        var transferAmount = BigInteger.ONE;
        Wallet receiver = KeyWallet.create();
        _transferTokenAndCheckResult(irc2Deployer, irc2.getAddress(), receiver.getAddress(), transferAmount, false);
        assertEquals(BigInteger.ZERO, irc2.balanceOf(receiver.getAddress()));

        _transferTokenAndCheckResult(sfOwner, irc2.getAddress(), receiver.getAddress(), transferAmount, true);
        RpcObject afterOutflow = sfScore.getOutflowInUSDT();
        assertEquals(transferAmount, irc2.balanceOf(receiver.getAddress()));
        final String CUSTOM = SustainableFundScore.OUTFLOW_CUSTOM;
        assertEquals(outflow.getItem(CUSTOM).asInteger().add(transferAmount),
                afterOutflow.getItem(CUSTOM).asInteger());
    }

    @Test
    void checkOutflowInUSDT() throws Exception {
        // check getOutflowInUSDT
        RpcObject object = sfScore.getInflowInUSDT();
        var irc2Deployer = wallets[1];
        IRC2TokenScore irc2 = IRC2TokenScore.mustDeploy(txHandler, irc2Deployer);
        _setUSDT(sfOwner, irc2.getAddress(), true);
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
