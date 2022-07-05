package io.havah.test.score;

import example.IRC2BasicToken;
import example.IRC3BasicToken;
import example.token.IRC2;
import example.token.IRC2Basic;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.ResultTimeoutException;
import foundation.icon.test.common.TransactionFailureException;
import foundation.icon.test.common.TransactionHandler;
import foundation.icon.test.score.Score;

import java.io.IOException;
import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;

public class IRC2TokenScore extends Score {
    private static final BigInteger MINT_STEP = BigInteger.valueOf(400000);
    private static final BigInteger TRANSFER_STEP = BigInteger.valueOf(500000);

    public IRC2TokenScore(Score other) {
        super(other);
    }

    public static IRC2TokenScore mustDeploy(TransactionHandler txHandler, Wallet owner)
            throws ResultTimeoutException, TransactionFailureException, IOException {
        var tokenClass = new Class[] {
                IRC2BasicToken.class,
                IRC2Basic.class,
                IRC2.class,
        };
        LOG.infoEntering("deploy", tokenClass[0].getName());
        RpcObject params = new RpcObject.Builder()
                .put("_name", new RpcValue("IRC2 token"))
                .put("_symbol", new RpcValue("RC2T"))
                .put("_decimals", new RpcValue("12"))
                .put("_initialSupply", new RpcValue("100"))
                .build();
        Score score = txHandler.deploy(owner, tokenClass, params);
        LOG.info("scoreAddr = " + score.getAddress());
        LOG.infoExiting();
        return new IRC2TokenScore(score);
    }

    public BigInteger balanceOf(Address owner) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_owner", new RpcValue(owner))
                .build();
        return call("balanceOf", params).asInteger();
    }

    public BigInteger totalSupply() throws IOException {
        return call("totalSupply", null).asInteger();
    }

    public Bytes transfer(Wallet wallet, Address to, BigInteger value) throws IOException {
        RpcObject params = new RpcObject.Builder()
                .put("_to", new RpcValue(to))
                .put("_value", new RpcValue(value))
                .build();
        return invoke(wallet, "transfer", params, null, TRANSFER_STEP);
    }
}

