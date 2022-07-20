package io.havah.test.common;

import foundation.icon.icx.data.Address;

import java.math.BigInteger;

public class Constants extends foundation.icon.test.common.Constants {
    public static final String TAG_HAVAH = "havah";

    public static final Address CHAINSCORE_ADDRESS
            = new Address("cx0000000000000000000000000000000000000000");
    public static final Address GOV_ADDRESS
            = new Address("cx0000000000000000000000000000000000000001");

    public static final Address SYSTEM_TREASURY
            = new Address("hx1000000000000000000000000000000000000000");

    public static final Address SUSTAINABLEFUND_ADDRESS
            = new Address("cx4000000000000000000000000000000000000000");

    public static final Address HOOVERFUND_ADDRESS
            = new Address("hx6000000000000000000000000000000000000000");

    public static final Address ECOSYSTEM_ADDRESS
            = new Address("cx7000000000000000000000000000000000000000");

    public static final Address PLANETNFT_ADDRESS
            = new Address("cx8000000000000000000000000000000000000000");

    public static final Address SERVICE_TREASURY
            = new Address("hx9000000000000000000000000000000000000000");

    public static final Address VAULT_ADDRESS
            = new Address("cx1100000000000000000000000000000000000000");

    public static final int RPC_ERROR_INVALID_ID = -30032;

    public static final BigInteger INITIAL_ISSUE_AMOUNT = new BigInteger("4300000000000000000000000", 10); // 초기 보상

    public static final BigInteger PRIVATE_LOCKUP = BigInteger.valueOf(2);

    public static final BigInteger PRIVATE_RELEASE_CYCLE = BigInteger.valueOf(4);

    public static BigInteger TOTAL_REWARD_PER_DAY
            = BigInteger.valueOf(430).multiply(BigInteger.TEN.pow(4)).multiply(BigInteger.TEN.pow(18));

    public static BigInteger HOOVER_BUDGET
            = TOTAL_REWARD_PER_DAY;

    public static BigInteger ECOSYSTEM_INITIAL_BALANCE
            = BigInteger.valueOf(15).multiply(BigInteger.TEN.pow(8)).multiply(BigInteger.TEN.pow(18));
}
