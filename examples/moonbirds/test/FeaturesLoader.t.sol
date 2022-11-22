// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "forge-std/Test.sol";
import "forge-std/console2.sol";

import "openzeppelin-contracts/utils/cryptography/MerkleProof.sol";

import "./TestLib.sol";
import "../script/FeaturesLoader.sol";

import "moonbirds-inchain/gen/Features.sol";

contract FeaturesLoaderTest is Test, FeaturesLoader {
    using TestLib for Vm;
    using FeaturesLib for Features;

    function testSerialise() public {
        assertEq(
            Features({
                background: 1,
                beak: 2,
                body: 3,
                eyes: 4,
                eyewear: 5,
                headwear: 6,
                outerwear: 7
            }).serialise(),
            0x01020304050607
        );
    }

    function _testProof(uint256 iBirb) internal {
        bytes32[] memory proof = getProof(iBirb);
        bool success = MerkleProof.verify(
            proof, FeaturesLib.FEATURES_ROOT, getFeatures(iBirb).hash(iBirb)
        );
        assertTrue(success);
    }

    function testAllProofs() public {
        for (uint256 i; i < 10_000; ++i) {
            _testProof(i);
        }
    }

    function testInvalidProof(uint256 iBirb) public {
        iBirb = bound(iBirb, 0, 9998);

        bytes32[] memory proof = getProof(iBirb);
        bool success = MerkleProof.verify(
            proof,
            FeaturesLib.FEATURES_ROOT,
            getFeatures(iBirb + 1).hash(iBirb + 1)
        );
        assertFalse(success);
    }
}
