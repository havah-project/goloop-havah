package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.http.HttpProvider;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.test.common.Env;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.ChainScore;
import io.havah.test.score.GovScore;
import io.havah.test.score.PlanetNFTScore;
import io.havah.test.score.SustainableFundScore;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

@Tag(Constants.TAG_HAVAH)
public class HooverFundTest extends TestBase {
    static PlanetNFTScore planetNFTScore;
    static SustainableFundScore sfScore;
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
        sfScore = new SustainableFundScore(txHandler);
        govScore = new GovScore(txHandler);
        chainScore = new ChainScore(txHandler);
        planetManager = KeyWallet.create();
        Wallet[] wallets = new Wallet[]{
                planetManager,
        };
        Utils.distributeCoin(wallets);

        var txResult = govScore.addPlanetManager(governor, planetManager.getAddress());
        assertEquals(BigInteger.ONE, txResult.getStatus());
    }

    static BigInteger _usdtAmountToGetGuaranteed(BigInteger guaranteed) {
        return guaranteed.multiply(BigInteger.valueOf(3600)).divide(BigInteger.TEN.pow(12));
    }

    static void _mintPlanet(Address holder, BigInteger usdt, BigInteger hvh) throws Exception {
        Bytes txHash = planetNFTScore.mintPlanet(governor, holder, 2, usdt, hvh);
        TransactionResult result = planetNFTScore.getResult(txHash);
        assertEquals(BigInteger.ONE, result.getStatus(), "failure result(" + result + ")");
    }

    BigInteger _chargeableHoover(BigInteger hooverBalance) throws Exception {
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

    private List<BigInteger> _tokenIds(Address holder) throws IOException {
        return planetNFTScore.tokenIdsOf(holder, 0, 50).tokenIds;
    }

    /*
        1. calculate guaranteed value
        2. minting
        3. wait util start of next term
        4. send reportPlanetWork
        5. claim & check that the values of guaranteed
     */
    private BigInteger _mintAndCheckReward(Wallet holder, BigInteger guaranteed) throws Exception {
        // 1. calculate guaranteed value
        var usdt = _usdtAmountToGetGuaranteed(guaranteed);
        guaranteed = usdt.multiply(BigInteger.TEN.pow(12)).divide(BigInteger.valueOf(3600));
        // 2.  mint
        _mintPlanet(holder.getAddress(), usdt, usdt.multiply(BigInteger.TEN.pow(12)));

        // check current hoover balance and check that hoover is refilled on next term
        var hooverBalance = _getBalance(Constants.HOOVERFUND_ADDRESS);
        var refilledAmount = _chargeableHoover(hooverBalance);
        Utils.waitUtilNextTerm();
        var cur = iconService.getLastBlock().execute().getHeight();
        Utils.waitUtil(cur.add(BigInteger.ONE));
        assertEquals(hooverBalance.add(refilledAmount), _getBalance(Constants.HOOVERFUND_ADDRESS),
                String.format("HF balance(%s), SF balance(%s)",
                        _getBalance(Constants.HOOVERFUND_ADDRESS),
                        _getBalance(Constants.SUSTAINABLEFUND_ADDRESS)));

        // check
        var tokenId = _tokenIds(holder.getAddress()).get(0);
        var rpwResult = chainScore.reportPlanetWork(planetManager, tokenId);
        assertEquals(BigInteger.ONE, rpwResult.getStatus(), rpwResult.toString());

        var holderBalance = _getBalance(holder.getAddress());
        var txResult = chainScore.claimPlanetReward(holder, new BigInteger[]{tokenId});
        var fee = txResult.getStepPrice().multiply(txResult.getStepUsed());
        assertEquals(BigInteger.ONE, txResult.getStatus());
        BigInteger supply = planetNFTScore.totalSupply();
        var dailyReward = Constants.TOTAL_REWARD_PER_DAY.divide(supply);
        if (guaranteed.compareTo(dailyReward) > 0) {
            assertEquals(holderBalance.subtract(fee).add(guaranteed), _getBalance(holder.getAddress()),
                    "guaranteed(" + guaranteed + "), dailyReward(" + dailyReward + "), reportPlanetWork result(" + rpwResult + ")," +
                            "HF balance(" + _getBalance(Constants.HOOVERFUND_ADDRESS) + "), SF balance(" + _getBalance(Constants.SUSTAINABLEFUND_ADDRESS) + ")");
        }
        return tokenId;
    }

    /*
        소수점 계산으로 변경 시 나누기, 곱셈이 들어가면서 오차가 발생.
        실제 계산하여 발생할 보상 보장값 반환.
     */
    BigInteger _calcExpectedReward(BigInteger guaranteed, BigInteger usdtPrice) {
        var usdt = guaranteed.multiply(BigInteger.valueOf(3600)).divide(BigInteger.TEN.pow(12));
        return usdt.multiply(BigInteger.TEN.pow(12)).multiply(usdtPrice).divide(ONE_HVH).divide(BigInteger.valueOf(3600));
    }

    BigInteger _calcExpectedReward(BigInteger guaranteed) throws IOException {
        return _calcExpectedReward(guaranteed, chainScore.getUSDTPrice());
    }

    void printHooverBalance(String param) throws Exception {
        LOG.info(param + " HOOVER BALANCE(" + iconService.getBalance(Constants.HOOVERFUND_ADDRESS).execute() + ")");
    }

    // check after 5 term if reward is valid or not
    private void _checkNextTermsRewardAndHoover(HooverTokenInfo[] tokenInfo) throws Exception {
        var holderNum = tokenInfo.length;
        BigInteger supply = planetNFTScore.totalSupply();
        BigInteger dailyReward = Constants.TOTAL_REWARD_PER_DAY.divide(supply);
        Utils.waitUtilNextTerm();
        Utils.waitUtil(Utils.getHeightNext(1));

        final int TEST_TERM = 5;
        for (int i = 0; i < TEST_TERM; i++) {
            printHooverBalance("start of term");
            List<BigInteger> balanceList = new ArrayList<>();
            // call reportPlanetWork and get balance
            for (int j = 0; j < holderNum; j++) {
                balanceList.add(iconService.getBalance(tokenInfo[j].wallet.getAddress()).execute());
            }

            var txResultList = _reportAndClaimPlanet(tokenInfo);

            for (int j = 0; j < holderNum; j++) {
                var reportResult = txResultList.get(j)[0];
                var claimResult = txResultList.get(j)[1];
                var txFee = claimResult.getStepPrice().multiply(claimResult.getStepUsed());
                var expectedReward = tokenInfo[j].expectedDailyReward;
                var holderAddress = tokenInfo[j].wallet.getAddress();
                if (expectedReward.compareTo(dailyReward) > 0) {
                    var expected = balanceList.get(j).subtract(txFee).add(expectedReward);
                    assertEquals(expected, _getBalance(holderAddress),
                            String.format("(%d), before balance(%s), after (%s), fee(%s), expected reward(%s), reportResult(%s)",
                                    j, balanceList.get(j), _getBalance(holderAddress), txFee,
                                    expectedReward, reportResult));
                } else {
                    var expected = balanceList.get(j).subtract(txFee).add(dailyReward);
                    var balance = _getBalance(holderAddress);
                    assertTrue(
                            expected.equals(balance) || expected.add(BigInteger.ONE).equals(balance),
                            String.format("(%d), before balance(%s), after (%s), fee(%s), expected reward(%s), reportResult(%s)",
                                    j, balanceList.get(j), _getBalance(holderAddress), txFee,
                                    expectedReward, reportResult));
                }
                LOG.info(String.format("dailyReward(%s), expected(%s), reportResult(%s)",
                        dailyReward, expectedReward, reportResult));
            }
            printHooverBalance("after claimPlanet");
            assertTrue(Constants.HOOVER_BUDGET.compareTo(_getBalance(Constants.HOOVERFUND_ADDRESS)) > 0);
            Utils.waitUtilNextTerm();
            Utils.waitUtil(Utils.getHeightNext(1));
            assertEquals(0, Constants.HOOVER_BUDGET.compareTo(_getBalance(Constants.HOOVERFUND_ADDRESS)),
                    String.format("HV balance(%s), SF balance(%s)", iconService.getBalance(Constants.HOOVERFUND_ADDRESS).execute(),
                            iconService.getBalance(Constants.SUSTAINABLEFUND_ADDRESS).execute()));
        }
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
        LOG.infoEntering("checkHooverSupport");
        var startHeight = Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(startHeight);
        final int holderNum = 5;
        HooverTokenInfo[] tokenInfo = new HooverTokenInfo[holderNum];
        Wallet[] holders = new Wallet[holderNum];
        for (int i = 0; i < holderNum; i++) {
            var info = new HooverTokenInfo();
            info.wallet = KeyWallet.create();
            tokenInfo[i] = info;
            holders[i] = info.wallet;
        }
        Utils.distributeCoin(holders);

        BigInteger supply = planetNFTScore.totalSupply();
        var desiredReward = new ArrayList<>(
                List.of(
                        BigInteger.ONE,
                        Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(2)))));
        desiredReward.addAll(Stream.of(3, 4, 5).map(
                        // total reward / (current supply + minting count) + ONE_HAVAH
                        (i) -> Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(i))).add(ONE_HVH))
                .collect(Collectors.toList()));

        for (int i = 0; i < desiredReward.size(); i++) {
            tokenInfo[i].expectedDailyReward = _calcExpectedReward(desiredReward.get(i));
        }

        for (int i = 0; i < holderNum; i++) {
            // mint planet and check first reward
            tokenInfo[i].tokenId = _mintAndCheckReward(holders[i], desiredReward.get(i));
        }

        // check after 5 term if reward is valid or not
        _checkNextTermsRewardAndHoover(tokenInfo);
        LOG.infoExiting();
    }

    /*
        daily reward는 planet 구매 시 지불한 USDT를 rewardPlanetWork호출 시
        HVH환율을 이용해 HVH로 변환하여 연 10%에 해당하는 HVH를 지급한다.
        연 10%에 해당하는 HVH가 planet reward만으로 채워지지 않을 경우 hoover에서 지원한다.
        단, 해당 planet 시 구매 시 지급한 HVH만큼 reward로 지급된 경우에는 hoover에서 더 이상 지원하지 않는다.
        1. hoover 지원을 받을 만큼 mint planet
        2. hoover 지원을 받는다.
        3. planet 구매 시 지급한 HVH만큼을 reward로 받았을 경우 더 이상 hoover로부터 지원받지 못하는 것 확인.
     */
    @Test
    void checkEndOfHooverSupport() throws Exception {
        var startHeight = Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(startHeight);
        BigInteger supply = planetNFTScore.totalSupply();
        var expectedPlanetReward = Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(5)));
        final int testNum = 5;
        BigInteger[] desiredReward = new BigInteger[testNum];
        Arrays.fill(desiredReward, expectedPlanetReward.add(ONE_HVH));
        Wallet[] holders = new Wallet[testNum];
        for (int i = 0; i < testNum; i++) {
            holders[i] = KeyWallet.create();
        }
        Utils.distributeCoin(holders);

        BigInteger[] expectedDailyReward = new BigInteger[desiredReward.length];
        for (int i = 0; i < desiredReward.length; i++) {
            expectedDailyReward[i] = _calcExpectedReward(desiredReward[i]);
            LOG.info("expectedDailyReward[" + i + "] = " + expectedDailyReward[i]);
        }

        List<BigInteger> tokenList = new ArrayList<>();
        List<Bytes> txList = new ArrayList<>();
        List<BigInteger> usdtList = new ArrayList<>();
        int index = 1;
        for (var reward : desiredReward) {
            usdtList.add(reward.multiply(BigInteger.valueOf(index++)));
        }
        for (int i = 0; i < testNum; i++) {
            // mint with guaranteed
            var usdt = _usdtAmountToGetGuaranteed(desiredReward[i]);
            // set HVH price to get hoover support util 5 times
            // TODO change below logic
            txList.add(planetNFTScore.mintPlanet(governor, holders[i].getAddress(), 2, usdt, usdtList.get(i)));
        }

        for (int i = 0; i < testNum; i++) {
            var result = txHandler.getResult(txList.get(i));
            assertEquals(BigInteger.ONE, result.getStatus());
            tokenList.add(planetNFTScore.tokenIdsOf(holders[i].getAddress(), 0, 1).tokenIds.get(0));
        }

        supply = planetNFTScore.totalSupply();
        var planetReward = Constants.TOTAL_REWARD_PER_DAY.divide(supply);
        // 0 ~ 5 will get reward as guaranteed, next will get reward only planet reward
        var accumulatedReward = new BigInteger[desiredReward.length];
        Arrays.fill(accumulatedReward, BigInteger.ZERO);
        while (true) {
            Utils.waitUtilNextTerm();
            // invoke
            txList.clear();
            for (int j = 0; j < testNum; j++) {
                txList.add(
                        chainScore.invokeReportPlanetWork(planetManager, tokenList.get(j)));
            }

            for (int j = 0; j < testNum; j++) {
                var txResult = txHandler.getResult(txList.get(j));
            }

            txList.clear();
            // claim
            List<BigInteger> beforeBalance = new ArrayList<>();
            for (int j = 0; j < testNum; j++) {
                beforeBalance.add(iconService.getBalance(holders[j].getAddress()).execute());
            }
            for (int j = 0; j < testNum; j++) {
                txList.add(chainScore.invokeClaimPlanetReward(holders[j], new BigInteger[]{tokenList.get(j)}));
            }

            for (int j = 0; j < testNum; j++) {
                var result = txHandler.getResult(txList.get(j));
                var txFee = result.getStepPrice().multiply(result.getStepUsed());
                beforeBalance.set(j, beforeBalance.get(j).subtract(txFee));
            }

            int receiveAllGuaranteed = 0;
            RpcObject obj = chainScore.getIssueInfo();
            var termSequence = obj.getItem("termSequence").asInteger();

            for (int j = 0; j < testNum; j++) {
                var balance = iconService.getBalance(holders[j].getAddress()).execute();
                var diff = balance.subtract(beforeBalance.get(j));
                if (accumulatedReward[j].compareTo(usdtList.get(j)) >= 0) {
                    LOG.info("NO HF support in " + termSequence + " for " + j);
                    receiveAllGuaranteed++;
                    assertTrue(
                            planetReward.equals(diff) ||
                                    planetReward.add(BigInteger.ONE).equals(diff),
                            "beforeBalance(" + beforeBalance + "), afterBalance(" + balance + "), reward(" + planetReward + "), diff(" + diff + ")");
                } else {
                    LOG.info("HF support in " + termSequence + " for " + j);
                    assertEquals(expectedDailyReward[j], diff
                            , "beforeBalance(" + beforeBalance + "), afterBalance(" + balance + "), reward(" + planetReward + "), expected(" + expectedDailyReward[j] + ")");
                }
                accumulatedReward[j] = accumulatedReward[j].add(diff);
            }
            if (receiveAllGuaranteed == testNum) {
                LOG.info("No more HF support in " + termSequence);
                break;
            }
        }
    }

    BigInteger _getBalance(Address address) throws Exception {
        return iconService.getBalance(address).execute();
    }

    private List<TransactionResult[]> _reportAndClaimPlanet(HooverTokenInfo[] tokenInfo) throws Exception {
        List<TransactionResult[]> txResultList = new ArrayList<>();
        List<Bytes> txHashList = new ArrayList<>();
        TransactionResult txResult;
        for (var info : tokenInfo) {
            txHashList.add(chainScore.invokeReportPlanetWork(planetManager, info.tokenId));
        }
        // getResult
        for (var hash : txHashList) {
            txResult = txHandler.getResult(hash);
            txResultList.add(new TransactionResult[]{txResult, null});
            assertEquals(BigInteger.ONE, txResult.getStatus());
        }
        txHashList.clear();
        for (var info : tokenInfo) {
            txHashList.add(chainScore.invokeClaimPlanetReward(info.wallet, new BigInteger[]{info.tokenId}));
        }
        for (var i = 0; i < txHashList.size(); i++) {
            txResult = txHandler.getResult(txHashList.get(i));
            txResultList.get(i)[1] = txResult;
            assertEquals(BigInteger.ONE, txResult.getStatus());
        }
        return txResultList;
    }

    private void _checkHooverRefill(HooverTokenInfo[] tokenInfo) throws Exception {
        final int TEST_TERM = 5;
        for (int i = 0; i < TEST_TERM; i++) {
            var sfBalance = _getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
            BigInteger inflowAmount = sfScore.getInflowAmount();
            var bInflow = sfScore.getInflow();
            _reportAndClaimPlanet(tokenInfo);
            var hfBalance = _getBalance(Constants.HOOVERFUND_ADDRESS);
            Utils.waitUtilNextTerm();
            Utils.waitUtil(Utils.getHeightNext(1));
            var aHooverBalance = _getBalance(Constants.HOOVERFUND_ADDRESS);
            var aSfBalance = _getBalance(Constants.SUSTAINABLEFUND_ADDRESS);
            var sfDiff = sfBalance.subtract(aSfBalance);
            var inflowDiff = sfScore.getInflowAmount().subtract(inflowAmount);
            LOG.info("SF inflowDiff(" + inflowDiff + "), sfDiff(" + sfDiff + ")");
            var refilled = inflowDiff.add(sfDiff);
            var log = "Loop(" + i + "), hoover : before(" + hfBalance + "), after(" + aHooverBalance + "), " +
                    "SF : before(" + sfBalance + "), after(" + aSfBalance + ")\n" +
                    "hoover increase(" + aHooverBalance.subtract(hfBalance) + "), " +
                    "sf decrease(" + sfBalance.subtract(aSfBalance) + "), " +
                    "before : inflowAmount(" + inflowAmount + "), after(" + sfScore.getInflowAmount() + "), " +
                    "inflowDiff(" + inflowDiff + "), " +
                    "before inflow(" + bInflow + "), inflow(" + sfScore.getInflow() + ")";
            assertEquals(aHooverBalance.subtract(hfBalance), refilled, log);
            LOG.info(log);
        }
    }

    /*
        SF에 HVH가 없을 때 HF에 얼마나 충전되는지 확인.
     */
    @Test
    void notEnoughHVHInSF() throws Exception {

        var originUsdtPrice = chainScore.getUSDTPrice();
        // 1 USDT = 10 HVH
        assertEquals(BigInteger.ONE,
                govScore.setUSDTPrice(governor, ONE_HVH.multiply(BigInteger.TEN))
                        .getStatus());
        // mint
        var mintingInfo = _mintPlanetForHooverTest();
        // transfer HVH from SF to one
        // check HOW to refill Hoover
        Utils.waitUtil(Utils.startRewardIssueIfNotStarted());

        Utils.waitUtilNextTerm();
        Utils.waitUtil(Utils.getHeightNext(1));
        // transfer SF balance to wallet
        Wallet tmpWallet = KeyWallet.create();
        var sfBalance = iconService.getBalance(Constants.SUSTAINABLEFUND_ADDRESS).execute();
        assertEquals(BigInteger.ONE,
                txHandler.getResult(
                                sfScore.transfer(governor, tmpWallet.getAddress(), sfBalance.subtract(BigInteger.valueOf(100))))
                        .getStatus());

        _checkHooverRefill(mintingInfo);

        txHandler.transfer(tmpWallet, Constants.SUSTAINABLEFUND_ADDRESS, sfBalance.subtract(BigInteger.valueOf(1000)));
        assertEquals(BigInteger.ONE,
                govScore.setUSDTPrice(governor, originUsdtPrice)
                        .getStatus());
    }

    /*
        USDT가격을 term별로 변화시키고 reward를 확인한다.
     */
    void _changeUSDTPriceAndCheckReward(HooverTokenInfo[] hooverTokenInfos, BigInteger[] usdtPrices,
                                        BigInteger[] desiredReward, BigInteger originUsdtPrice) throws Exception {
        List<Bytes> txHashList = new ArrayList<>();
        List<Bytes> reportTxHasList = new ArrayList<>();
        for (int i = 0; i <= usdtPrices.length; i++) {
            Utils.waitUtilNextTerm();
            TransactionResult result;
            if (i < usdtPrices.length) {
                result = govScore.setUSDTPrice(governor, usdtPrices[i]);
                assertEquals(BigInteger.ONE, result.getStatus());
            }

            for (var info : hooverTokenInfos) {
                reportTxHasList.add(chainScore.invokeReportPlanetWork(planetManager, info.tokenId));
            }
            for (int j = 0; j < reportTxHasList.size(); j++) {
                result = txHandler.getResult(reportTxHasList.get(j));
                assertEquals(BigInteger.ONE, result.getStatus(), "failure (" + i + "), result(" + result + ")");
                LOG.info("expected(" + hooverTokenInfos[j].expectedDailyReward + ")");
            }
            List<BigInteger> balances = new ArrayList<>();
            for (var info : hooverTokenInfos) {
                balances.add(_getBalance(info.wallet.getAddress()));
                txHashList.add(chainScore.invokeClaimPlanetReward(info.wallet, new BigInteger[]{info.tokenId}));
            }

            for (int j = 0; j < txHashList.size(); j++) {
                result = txHandler.getResult(txHashList.get(j));
                assertEquals(BigInteger.ONE, result.getStatus(), result.toString());
                var fee = result.getStepPrice().multiply(result.getStepUsed());
                LOG.info(String.format("expected(%s), index(%d)", hooverTokenInfos[j].expectedDailyReward, j));
                var expected = hooverTokenInfos[j].expectedDailyReward;
                var supply = planetNFTScore.totalSupply();
                var todayReward = Constants.TOTAL_REWARD_PER_DAY.divide(supply);
                if (!expected.equals(BigInteger.ZERO)) {
                    var hooverBalance = _getBalance(Constants.HOOVERFUND_ADDRESS);
                    if (!BigInteger.ZERO.equals(hooverBalance)) {
                        BigInteger expectedReward;
                        BigInteger usdtPrice;
                        if (i == 0) {
                            usdtPrice = originUsdtPrice;
                        } else {
                            usdtPrice = usdtPrices[i - 1];
                        }
                        expectedReward = _calcExpectedReward(desiredReward[j], usdtPrice);
                        expectedReward = expectedReward.compareTo(todayReward) > 0 ? expected : todayReward;
                        LOG.info("usdtPrice(" + usdtPrice + "), desiredReward(" + desiredReward[j] + "), expectedReward(" + expectedReward + ")");
                        var finalReward = balances.get(j).subtract(fee).add(expectedReward);
                        assertTrue(finalReward.equals(_getBalance(hooverTokenInfos[j].wallet.getAddress()))
                                        || finalReward.add(BigInteger.ONE).equals(_getBalance(hooverTokenInfos[j].wallet.getAddress())),
                                String.format("LOOP(%d), info(%s), reward before(%s), after(%s), today(%s), txFee(%s), hooverBalance(%s), expectedReward(%s), reportResult(%s)",
                                        i, hooverTokenInfos[j], balances.get(j), _getBalance(hooverTokenInfos[j].wallet.getAddress()),
                                        todayReward, fee, _getBalance(Constants.HOOVERFUND_ADDRESS), expectedReward,
                                        txHandler.getResult(reportTxHasList.get(j))));
                    }
                } else {
                    var expBalance = balances.get(j).add(todayReward).subtract(fee);
                    var curBalance = _getBalance(hooverTokenInfos[j].wallet.getAddress());
                    assertTrue(expBalance.equals(curBalance) || expBalance.add(BigInteger.ONE).equals(curBalance), // add reminder
                            String.format("reward before(%s), after(%s), today(%s), txFee(%s), hooverBalance(%s)", balances.get(j), _getBalance(hooverTokenInfos[j].wallet.getAddress()),
                                    todayReward, fee, _getBalance(Constants.HOOVERFUND_ADDRESS)));
                }
            }
            txHashList.clear();
            reportTxHasList.clear();
        }
    }

    /*
        USDT가격 변동에 따라 HOOVER에서 지급되는 HVH가 달라지는지 확인한다.
        1. mint Hoover가 필요한 금액으로 동일하게 3대 mint
        2. term 중간에 term이전에 변경하여 적용된 price에 맞게 hoover가 지급되는지 확인한다.
     */
    @Test
    void checkHooverPaymentsAccordingToUSDTPrice() throws Exception {
        Wallet notUsed = KeyWallet.create();
        // mint to refill SF.
        // these planets are not be working.
        for (int i = 0; i < 10; i++) {
            planetNFTScore.mintPlanet(governor, notUsed.getAddress(), 2, BigInteger.ONE, BigInteger.ONE);
        }
        var originUsdtPrice = chainScore.getUSDTPrice();
        var startHeight = Utils.startRewardIssueIfNotStarted();
        Utils.waitUtil(startHeight);
        // mint
        // 2, 10, 100
        BigInteger supply = planetNFTScore.totalSupply();
        BigInteger[] desiredReward = new BigInteger[]{
                BigInteger.ONE, // not hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.TWO)), // not hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(3))).add(ONE_HVH), // hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(4))).add(ONE_HVH.multiply(BigInteger.TWO)), // hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(5))).add(ONE_HVH.multiply(BigInteger.valueOf(3))), // hoover support
        };

        var hooverTokenInfos = _mintPlanetForHooverTest(desiredReward);
        var usdtPrices = new BigInteger[]{
                ONE_HVH.multiply(BigInteger.TEN), // 1 USDT = 10 HVH
                ONE_HVH.multiply(BigInteger.TEN).add(ONE_HVH.divide(BigInteger.TEN.pow(2))), // 1 USDT = 10.01 HVH
                ONE_HVH.multiply(BigInteger.TEN.multiply(BigInteger.TWO)), // 1 USDT = 20 HVH
                ONE_HVH.divide(BigInteger.TEN), // 1 USDT = 0.1 HVH
        };

        _changeUSDTPriceAndCheckReward(hooverTokenInfos, usdtPrices, desiredReward, originUsdtPrice);
        assertEquals(BigInteger.ONE, govScore.setUSDTPrice(governor, originUsdtPrice).getStatus());
    }

    /*
        Hoover를 모두 사용했을 때 다른 planet에게 reward를 주지 못하는 지 확인한다.
     */
    @Test
    void spendAllHVHInHoover() throws Exception {
        // transfer hoover to one

        // down HVH price 1/10000

        var tokenInfo = _mintPlanetForHooverTest();
    }

    /*
        generate wallets for test
        mint planet with desiredReward to wallets
        return wallet, tokenId,
     */
    HooverTokenInfo[] _mintPlanetForHooverTest(BigInteger[] desiredReward) throws Exception {
        // generate wallets
        final int testNum = desiredReward.length;
        HooverTokenInfo[] tokenInfo = new HooverTokenInfo[testNum];
        Wallet[] holders = new Wallet[testNum];
        for (int i = 0; i < testNum; i++) {
            holders[i] = KeyWallet.create();
            var hooverToken = new HooverTokenInfo();
            hooverToken.wallet = holders[i];
            tokenInfo[i] = hooverToken;
        }
        Utils.distributeCoin(holders);

        // estimate daily guaranteed reward
        BigInteger[] expectedDailyReward = new BigInteger[desiredReward.length];
        for (int i = 0; i < desiredReward.length; i++) {
            expectedDailyReward[i] = _calcExpectedReward(desiredReward[i]);
            tokenInfo[i].expectedDailyReward = _calcExpectedReward(desiredReward[i]);
        }

        // mint planet with desired reward
        List<Bytes> txList = new ArrayList<>();
        for (int i = 0; i < testNum; i++) {
            //  mint with guaranteed
            var usdt = _usdtAmountToGetGuaranteed(desiredReward[i]);
            // set HVH price to get hoover support util 5 times
            txList.add(planetNFTScore.mintPlanet(governor, holders[i].getAddress(), 2, usdt, expectedDailyReward[i].multiply(BigInteger.valueOf(1000))));
        }

        for (int i = 0; i < testNum; i++) {
            var result = txHandler.getResult(txList.get(i));
            assertEquals(BigInteger.ONE, result.getStatus());
            tokenInfo[i].tokenId = _tokenIds(holders[i].getAddress()).get(0);
        }
        return tokenInfo;
    }

    HooverTokenInfo[] _mintPlanetForHooverTest() throws Exception {
        BigInteger supply = planetNFTScore.totalSupply();
        BigInteger[] desiredReward = new BigInteger[]{
                BigInteger.ONE, // not hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.TWO)), // not hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(3))).add(ONE_HVH), // hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(4))).add(ONE_HVH), // hoover support
                Constants.TOTAL_REWARD_PER_DAY.divide(supply.add(BigInteger.valueOf(5))).add(ONE_HVH), // hoover support
        };
        return _mintPlanetForHooverTest(desiredReward);
    }

    public static class HooverTokenInfo {
        Wallet wallet;
        BigInteger expectedDailyReward;
        BigInteger tokenId;

        @Override
        public String toString() {
            return "HooverTokenInfo{" +
                    "wallet=" + wallet +
                    ", expectedDailyReward=" + expectedDailyReward +
                    ", tokenId=" + tokenId +
                    '}';
        }
    }
}