package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.http.HttpProvider;
import foundation.icon.test.common.Env;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.ChainScore;
import io.havah.test.score.GovScore;
import io.havah.test.score.PlanetNFTScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

@Tag(Constants.TAG_HAVAH)
public class HooverFundTest extends TestBase {
    static PlanetNFTScore planetNFTScore;
    static TransactionHandler txHandler;
    private static IconService iconService;
    private static ChainScore chainScore;
    private static GovScore govScore;
    private static Wallet governor;
    private static Wallet planetManager;
    private static final BigInteger ONE_HVH = BigInteger.ONE.multiply(BigInteger.TEN.pow(18));

    /*
    1. planet 구매 with 매우 높은 USDT구매가격 필요.
    2. reportPlanetWork호출 시 hoover fund에서 자금 조달되는지 확인.
    3. 다음 Term에서 hoover fund에 값이 채워지는지 확인.
    4. 채워진 값은 sustainable fund에서 전달되었는지 확인.
    5. sustainable fund의 outflow값이 변경되는지 확인.
     */
    @BeforeAll
    static void setup() throws Exception {
        Env.Channel channel = Env.nodes[0].channels[0];
        Env.Chain chain = channel.chain;
        iconService = new IconService(new HttpProvider(channel.getAPIUrl(Env.testApiVer)));
        txHandler = new TransactionHandler(iconService, chain);
        governor = txHandler.getChain().governorWallet;
        planetNFTScore = new PlanetNFTScore(governor, txHandler);
        govScore = new GovScore(txHandler);
        chainScore = new ChainScore(txHandler);
        planetManager = KeyWallet.create();
        // 5 * 10^13 + 1*10^7
        // USDT * rate
        // 이자 = 구매 USDT가격 * USDT 가격 * 1/360 * 1/10
        // 구매 USDT 가격 = 이자 / USDT가격 * 360 * 10
        // hoover 받기 위해
        // 이자 = 5백만 HVH / (totalSuuply+1) + A => A만큼 hoover에서 전달.
        //
//        BigInteger usdt = BigInteger.valueOf(5).multiply(BigInteger.TEN.pow(12)).add(BigInteger.TEN.pow(7)).multiply(BigInteger.valueOf(36010));
//        BigInteger usdt = BigInteger.valueOf(5).multiply(BigInteger.TEN.pow(12)).add(BigInteger.TEN.pow(7));
//        BigInteger usdt = BigInteger.TEN;
        // 5 * 000, 000 * 10^18
//        BigInteger usdt = _reward(BigInteger.valueOf(5).multiply(BigInteger.TEN.pow(24)));
//        usdt.add(BigInteger.ONE);
        Wallet[] wallets = new Wallet[]{
                governor,
                planetManager,
        };

        Utils.distributeCoin(wallets);

        var usdtPrice = chainScore.getUSDTPrice();
        BigInteger supply = planetNFTScore.totalSupply();
        LOG.info("supply(" + supply + "), usdtPrice(" + usdtPrice + ")");
        BigInteger usdt = BigInteger.valueOf(5).multiply(BigInteger.TEN.pow(12)).multiply(BigInteger.valueOf(3600)).divide(supply.add(BigInteger.ONE));
//        BigInteger usdt1 = usdt.add(BigInteger.ONE);
        LOG.info("USDT(" + usdt + ")");
//        LOG.info("expected hoover(" + usdt.subtract(usdt1) + ")");
//        _mintPlanet(holder.getAddress(), usdt.add(BigInteger.ONE), usdt.multiply(BigInteger.TEN.pow(13)));
//        _prepareReward(planetManager.getAddress());
        var txResult = govScore.addPlanetManager(governor, planetManager.getAddress());
    }

    static BigInteger _usdtAmountToGetGuaranteed(BigInteger guaranteed) {
        return guaranteed.multiply(BigInteger.valueOf(3600)).divide(BigInteger.TEN.pow(12));
//        return guaranteed / USDT price * 360 * 10;
    }

    static BigInteger _reward(BigInteger guaranteed) {
        return guaranteed.multiply(BigInteger.valueOf(3600)).divide(BigInteger.TEN.pow(12));
    }

    static void _mintPlanet(Address holder, BigInteger usdt, BigInteger hvh) throws Exception {
        Bytes txHash = planetNFTScore.mintPlanet(governor, holder, 2, usdt, hvh);
        TransactionResult result = planetNFTScore.getResult(txHash);
        assertEquals(BigInteger.ONE, result.getStatus(), "failure result(" + result + ")");
    }

    static void _prepareReward(Address planetManager) throws Exception {
        var txResult = govScore.addPlanetManager(governor, planetManager);
        var issueInfo = chainScore.getIssueInfo();
        var issueStart = issueInfo.getItem("issueStart");
        BigInteger issueStartHeight;
        LOG.info("issueInfo0(" + chainScore.getIssueInfo() + ")");
        if (issueStart == null) {
            // start and wait
            var height = Utils.getHeight();
            issueStartHeight = height.add(BigInteger.valueOf(3));
            govScore.startRewardIssue(governor, issueStartHeight);
        } else {
            issueStartHeight = issueStart.asInteger();
        }
        Utils.waitUtil(issueStartHeight);
    }

    // HVH to USDT
    BigInteger convertToUsdt(BigInteger value) throws Exception {
        return value.divide(chainScore.getUSDTPrice().divide(ONE_HVH));
    }

    BigInteger refillableAmount(BigInteger hooverBalance) throws Exception {
        var need = Constants.HOOVER_BUDGET.subtract(hooverBalance);
        var sfBalance = iconService.getBalance(Constants.SUSTAINABLEFUND_ADDRESS).execute();
        var refillAmount = BigInteger.ZERO;
        if (sfBalance.compareTo(need) >= 0) {
            refillAmount = need;
        } else {
            refillAmount = sfBalance;
        }
        return refillAmount;
    }

    /*
        1. minting
        2. wait util start of next term
        3. send reportPlanetWork
        4. claim
        5. check that the values of guaranteed and claimed value match.
     */
    private BigInteger _mintAndCheckReward(Wallet holder, BigInteger guaranteed) throws Exception {
        var usdt = _usdtAmountToGetGuaranteed(guaranteed);
        guaranteed = usdt.multiply(BigInteger.TEN.pow(12)).divide(BigInteger.valueOf(3600));
        _mintPlanet(holder.getAddress(), usdt, usdt.multiply(BigInteger.TEN.pow(12)));
        var hooverBalance = iconService.getBalance(Constants.HOOVERFUND_ADDRESS).execute();
        var refillAmount = refillableAmount(hooverBalance);
        Utils.waitUtilNextTerm();
        assertEquals(hooverBalance.add(refillAmount), iconService.getBalance(Constants.HOOVERFUND_ADDRESS).execute());

        // check balance of S0x12cF
        // check
        var tokenInfo = planetNFTScore.tokenIdsOf(holder.getAddress(), 0, 1);
        var tokenId = tokenInfo.tokenIds.get(0);
        var rpwResult = chainScore.reportPlanetWork(planetManager, tokenId);
        assertEquals(BigInteger.ONE, rpwResult.getStatus(), rpwResult.toString());

        var holderBalance = iconService.getBalance(holder.getAddress()).execute();
        var txResult = chainScore.claimPlanetReward(holder, new BigInteger[]{tokenId});
        var fee = txResult.getStepPrice().multiply(txResult.getStepUsed());
        assertEquals(BigInteger.ONE, txResult.getStatus());
        BigInteger supply = planetNFTScore.totalSupply();
        var dailyReward = Constants.TOTAL_REWARD_PER_DAY.divide(supply);
        if (guaranteed.compareTo(dailyReward) > 0) {
            assertEquals(holderBalance.subtract(fee).add(guaranteed), iconService.getBalance(holder.getAddress()).execute(),
                    "guaranteed(" + guaranteed + "), reportPlanetWork result(" + rpwResult + ")");
        }
        return tokenId;
    }

    BigInteger _increasePerDay(BigInteger inHVH) {
        return inHVH.multiply(BigInteger.valueOf(3600)).multiply(BigInteger.TEN.pow(18));
    }

    /*
        소수점 계산으로 변경 시 나누기, 곱셈이 들어가면서 오차가 발생.
        실제 계산하여 발생할 보상 보장값 반환.
     */
    BigInteger _correctNotMatchedValue(BigInteger guaranteed) {
        var usdt = guaranteed.multiply(BigInteger.valueOf(3600)).divide(BigInteger.TEN.pow(12));
        return usdt.multiply(BigInteger.TEN.pow(12)).divide(BigInteger.valueOf(3600));
    }

    /*
        1. hoover가 사용되도록 보장 금액 설정 & USDT값으로 변환.
        2. hoover에서 보존되는 금액 확인.
        3. hoover balance 확인.
            - planet별 하루 보상양이 나머지에 의해 약간의 오차가 발생할 수 있음. ex) 5 / 3 => 나머지 2.
            - 나머지는 계속 모았다가 분배시 몫이 생기면 planet별 보상양이 조금 증가하명서 hoover 보존금액이 줄어들 수 있음.
            - startRewarIssue부터 나머지 수를 모두 추적해서 계산에 담는것은 불가능하여 테스트 케이스에서 제외.
        4. term 이후 hoover refill & SF 출금 확인
     */
    @Test
    void checkHooverSupport() throws Exception {
        var startHeight = Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(startHeight);
        final int holderNum = 5;
        Wallet[] holders = new Wallet[holderNum];
        for (int i = 0; i < holderNum; i++) {
            holders[i] = KeyWallet.create();
        }
        Utils.distributeCoin(holders);

        BigInteger supply = planetNFTScore.totalSupply();
        BigInteger[] guaranteed = new BigInteger[]{
                BigInteger.ONE, // not hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.TWO)), // not hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(3))).add(ONE_HVH), // hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(4))).add(ONE_HVH), // hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(5))).add(ONE_HVH), // hoover support
        };
        BigInteger[] corrected = new BigInteger[guaranteed.length];
        for (int i = 0; i < guaranteed.length; i++) {
            corrected[i] = _correctNotMatchedValue(guaranteed[i]);
        }

        // mint
        List<BigInteger> tokenList = new ArrayList<>();
        for (int i = 0; i < holderNum; i++) {
            tokenList.add(_mintAndCheckReward(holders[i], guaranteed[i]));
        }

        for (int j = 0; j < 5; j++) {
            var hfBalance = iconService.getBalance(Constants.HOOVERFUND_ADDRESS).execute();
            var sfBalance = iconService.getBalance(Constants.SUSTAINABLEFUND_ADDRESS).execute();
            Utils.waitUtilNextTerm();
            assertEquals(Constants.HOOVER_BUDGET, iconService.getBalance(Constants.HOOVERFUND_ADDRESS).execute(), "hoover(" + hfBalance + "), sf(" + sfBalance + ")");
            List<Bytes> hashList = new ArrayList<>();
            List<BigInteger> balList = new ArrayList<>();
            for (int i = 0; i < holderNum; i++) {
                hashList.add(
                        chainScore.invokeReportPlanetWork(planetManager, tokenList.get(i)));
                balList.add(iconService.getBalance(holders[i].getAddress()).execute());
            }
            for (int i = 0; i < holderNum; i++) {
                var result = iconService.getTransactionResult(hashList.get(i)).execute();
                assertEquals(BigInteger.ONE, result.getStatus());
            }
            hashList.clear();
            for (int i = 0; i < holderNum; i++) {
                hashList.add(
                        chainScore.invokeClaimPlanetReward(planetManager, new BigInteger[]{tokenList.get(i)}));
            }
            for (int i = 0; i < holderNum; i++) {
                var result = iconService.getTransactionResult(hashList.get(i)).execute();
                assertEquals(BigInteger.ONE, result.getStatus());
                var txFee = result.getStepPrice().multiply(result.getStepUsed());
                var expected = balList.get(i).subtract(txFee).add(corrected[i]);
                assertEquals(expected, iconService.getBalance(holders[i].getAddress()).execute());
            }
            assertTrue(Constants.HOOVER_BUDGET.compareTo(iconService.getBalance(Constants.HOOVERFUND_ADDRESS).execute()) > 0);
            Utils.waitUtilNextTerm();
            assertEquals(0, Constants.HOOVER_BUDGET.compareTo(iconService.getBalance(Constants.HOOVERFUND_ADDRESS).execute()));
        }
    }

    @Test
    void checkEndOfHooverSupport() throws Exception {

    }
}