// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "forge-std/console2.sol";
import "forge-std/Test.sol";

import {IndexedBucketLib} from "solidify-contracts/IndexedBucketLib.sol";

contract IndexedBucketLibTest is Test {
    using IndexedBucketLib for bytes;

    function setUp() public {}

    function _getIndexedFields(bytes[] memory fields)
        internal
        pure
        returns (bytes memory ret)
    {
        uint16 loc = uint16(fields.length) * 2;
        for (uint256 i; i < fields.length; ++i) {
            ret = abi.encodePacked(ret, loc);
            loc += uint16(fields[i].length);
        }
        for (uint256 i; i < fields.length; ++i) {
            ret = abi.encodePacked(ret, fields[i]);
        }
    }

    function testGetFirstField() public {
        // single field with [0xabcd]
        bytes memory data = hex"0002abcd";
        bytes memory field = data.getField(0);

        assertEq(field.length, 2);
        assertEq(field, hex"abcd");
    }

    function testGetSecondField() public {
        bytes[] memory fields = new bytes[](2);
        fields[0] = hex"abcd";
        fields[1] = hex"0123456789abcdef0123456789abcdef";

        assertEq(
            _getIndexedFields(fields),
            hex"00040006abcd0123456789abcdef0123456789abcdef"
        );

        assertEq(_getIndexedFields(fields).getField(0), fields[0]);
        assertEq(_getIndexedFields(fields).getField(1), fields[1]);
    }
}
