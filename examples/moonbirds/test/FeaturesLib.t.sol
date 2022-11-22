// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "./TestLib.sol";
import "forge-std/console2.sol";

import {FeaturesLib} from "moonbirds-inchain/gen/Features.sol";

contract FeaturesLibTest is MBTest {
    using FeaturesLib for Features;

    function setUp() public {}

    function testSerialiseRound() public {
        Features memory f = Features({
            background: 7,
            beak: 6,
            body: 5,
            eyes: 4,
            eyewear: 3,
            headwear: 2,
            outerwear: 1
        });

        assertEq(f, FeaturesLib.deserialise(f.serialise()));
    }

    function testDeserialiseFromBytesRound() public {
        Features memory f = Features({
            background: 7,
            beak: 6,
            body: 5,
            eyes: 4,
            eyewear: 3,
            headwear: 2,
            outerwear: 1
        });

        bytes memory data = hex"07060504030201";
        assertEq(f, FeaturesLib.deserialise(data));
    }
}
