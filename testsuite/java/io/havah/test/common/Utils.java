package io.havah.test.common;

import foundation.icon.icx.IconService;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Block;
import foundation.icon.icx.transport.http.HttpProvider;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.test.common.Constants;
import foundation.icon.test.common.Env;
import foundation.icon.test.common.ResultTimeoutException;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.score.ChainScore;
import io.havah.test.score.GovScore;

import java.io.IOException;
import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;

public final class Utils {
    private final static ChainScore _chainScore;
    private final static GovScore _govScore;
    private final static IconService _iconService;
    private final static TransactionHandler _txHandler;
    private static Wallet _governor;

    static {
        Env.Channel channel = Env.nodes[0].channels[0];
        Env.Chain chain = channel.chain;
        IconService iconService = new IconService(new HttpProvider(channel.getAPIUrl(Env.testApiVer)));
        TransactionHandler txHandler = new TransactionHandler(iconService, chain);
        _iconService = iconService;
        _chainScore = new ChainScore(txHandler);
        _txHandler = txHandler;
        _govScore = new GovScore(txHandler);
        _governor = txHandler.getChain().governorWallet;
    }

    public static Wallet getGovernor() {
        return _governor;
    }

    public static IconService getIconService() {
        return _iconService;
    }

    public static TransactionHandler getTxHandler() {
        return _txHandler;
    }

    public static BigInteger getHeight() throws IOException {
        Block lastBlk = _iconService.getLastBlock().execute();
        return lastBlk.getHeight();
    }

    public static void waitUtil(BigInteger height) throws Exception {
        var now = getHeight();
        while (now.compareTo(height) < 0) {
            LOG.info("now(" + now + "), wait(" + height + ")");
            Thread.sleep(1200);
            now = getHeight();
        }
    }

    // termPeriod - ((current height - issueStart) % termPeriod)
    public static BigInteger getHeightUntilNextTerm() throws Exception {
        var issueInfo = _chainScore.getIssueInfo();
        var termPeriod = issueInfo.getItem("termPeriod").asInteger();
        var height = issueInfo.getItem("height").asInteger();
        BigInteger issueStart;
        try {
            LOG.info("termSequence : " + issueInfo.getItem("termSequence").asInteger());
            issueStart = issueInfo.getItem("issueStart").asInteger();
            LOG.info("issueStart : " + issueStart);
        } catch (NullPointerException e) {
            LOG.info("startRewardIssue not called. termSequence and issueStart is null.");
            return null;
        }
        return height.add(termPeriod.subtract(getHeight().subtract(issueStart).remainder(termPeriod)));
    }

    public static void waitUtilNextTerm() throws Exception {
        var nextTerm = getHeightUntilNextTerm();
        waitUtil(nextTerm);
    }

    public static void distributeCoin(Wallet[] to, BigInteger value) throws Exception {
        var amount = value.multiply(BigInteger.TEN.pow(18));
        BigInteger[] oldVal = new BigInteger[to.length];
        for (int i = 0; i < to.length; i++) {
            oldVal[i] =  _txHandler.getBalance(to[i].getAddress());
            _txHandler.transfer(to[i].getAddress(), amount);
        }
        for (int i = 0; i < to.length; i++) {
            ensureIcxBalance(_txHandler, to[i].getAddress(), oldVal[i], amount);
        }
    }

    public static void distributeCoin(Wallet[] to) throws Exception {
        distributeCoin(to, BigInteger.valueOf(50));
    }

    private static void ensureIcxBalance(TransactionHandler txHandler, Address address,
                                           BigInteger oldVal, BigInteger newVal) throws Exception {
        long limitTime = System.currentTimeMillis() + Constants.DEFAULT_WAITING_TIME;
        while (true) {
            BigInteger icxBalance = txHandler.getBalance(address);
            String msg = "ICX balance of " + address + ": " + icxBalance + ", old balance " + oldVal;
            if (icxBalance.equals(oldVal)) {
                if (limitTime < System.currentTimeMillis()) {
                    throw new ResultTimeoutException();
                }
                try {
                    // wait until block confirmation
                    LOG.debug(msg + "; Retry in 1 sec.");
                    Thread.sleep(1000);
                } catch (InterruptedException e) {
                    e.printStackTrace();
                }
            } else if (icxBalance.equals(oldVal.add(newVal))) {
                LOG.info(msg);
                break;
            } else {
                throw new IOException(String.format("ICX balance mismatch: expected <%s>, but was <%s>",
                        newVal, icxBalance));
            }
        }
    }

    // return null if not called startRewardIssue
    // return issueStart if called startRewardIssue
    public static BigInteger startRewardIssueIfNotStarted() throws Exception {
        RpcObject obj = _chainScore.getIssueInfo();
        var termSequence = obj.getItem("termSequence");
        if (termSequence == null) {
            var startHeight = getHeight().add(BigInteger.valueOf(5));
            _govScore.startRewardIssue(_governor, startHeight);
            return startHeight;
        }
        return termSequence.asInteger();
    }
}
