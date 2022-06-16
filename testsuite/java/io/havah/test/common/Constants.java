package io.havah.test.common;

import foundation.icon.icx.data.Address;

public class Constants extends foundation.icon.test.common.Constants {
    public static final String TAG_HAVAH = "havah";

    public static final Address CHAINSCORE_ADDRESS
            = new Address("cx0000000000000000000000000000000000000000");
    public static final Address GOV_ADDRESS
            = new Address("cx0000000000000000000000000000000000000001");
    public static final Address TREASURY_ADDRESS
            = new Address("hx1000000000000000000000000000000000000000");
    public static final Address PLANETNFT_ADDRESS
            = new Address("cx8000000000000000000000000000000000000000");

    public static final int RPC_ERROR_INVALID_ID = -30032;
}
