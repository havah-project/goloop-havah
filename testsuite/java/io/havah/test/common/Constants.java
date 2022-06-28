package io.havah.test.common;

import foundation.icon.icx.data.Address;

import java.math.BigInteger;

public class Constants extends foundation.icon.test.common.Constants {
    public static final String TAG_HAVAH = "havah";

    public static final Address CHAINSCORE_ADDRESS
            = new Address("cx0000000000000000000000000000000000000000");
    public static final Address GOV_ADDRESS
            = new Address("cx0000000000000000000000000000000000000001");

    public static final Address TREASURY_ADDRESS
            = new Address("hx1000000000000000000000000000000000000000");

    public static final Address SUSTAINABLEFUND_ADDRESS
            = new Address("cx4000000000000000000000000000000000000000");

    public static final Address HOOVERFUND_ADDRESS
            = new Address("hx6000000000000000000000000000000000000000");

    public static final Address ECOSYSTEM_ADDRESS
            = new Address("cx7000000000000000000000000000000000000000");

    public static final Address PLANETNFT_ADDRESS
            = new Address("cx8000000000000000000000000000000000000000");

    public static final int RPC_ERROR_INVALID_ID = -30032;

    public static final BigInteger INITIAL_ISSUE_AMOUNT = new BigInteger("5000000000000000000000000", 10); // 초기 발행량
}
