// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "forge-std/console2.sol";
import "forge-std/Test.sol";

import {LabelledBucketLib} from "solidify-contracts/LabelledBucketLib.sol";

contract LabelledBucketLibTest is Test {
    using LabelledBucketLib for bytes;

    function setUp() public {}

    function testFindFieldByLabel() public {
        _testFindFieldByLabel(
            hex"0001a10010a20011a30200a40201a50301a6", 0x0010, hex"a2"
        );
        _testFindFieldByLabel(
            hex"0001a10010a20011a30200a40201a5", 0x0200, hex"a4"
        );
    }

    function _testFindFieldByLabel(
        bytes memory data,
        uint16 label,
        bytes memory want
    ) internal {
        bytes memory field = data.findFieldByLabel(label, 1);
        assertEq(field.length, 1);
        assertEq(field, want);
    }

    function testCannotFindFieldByInvalidLabel() public {
        vm.expectRevert(
            abi.encodeWithSelector(
                LabelledBucketLib.LabelNotFound.selector, 0x0002
            )
        );
        _testFindFieldByLabel(hex"0001a10010a2", 0x0002, hex"00");
    }

    function testCannotFindFieldByLabelWithInvalidBounds() public {
        vm.expectRevert(
            abi.encodeWithSelector(
                LabelledBucketLib.InvalidBinarySearchBound.selector,
                0x0012,
                0x0001,
                0x0010
            )
        );
        _testFindFieldByLabel(hex"0001a10010a2", 0x0012, hex"00");
    }
}
