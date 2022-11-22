// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "forge-std/Test.sol";
import "forge-std/console2.sol";
import "./TestLib.sol";

import {FeaturesStorageManager} from
    "moonbirds-inchain/FeaturesStorageManager.sol";
import {FeaturesStorageDeployer} from
    "moonbirds-inchain/gen/FeaturesStorageDeployer.sol";

import {FeaturesStorageMapping} from
    "moonbirds-inchain/gen/FeaturesStorageMapping.sol";
import {FeaturesLoader} from "../script/FeaturesLoader.sol";

import {FeaturesLib} from "moonbirds-inchain/gen/Features.sol";

contract FeaturesStorageTest is MBTest, FeaturesLoader {
    using TestLib for Vm;
    using FeaturesLib for Features;

    FeaturesStorageManager public manager;

    function setUp() public {
        manager = new FeaturesStorageManager(
            FeaturesStorageDeployer.deployAsStatic()
        );
    }

    function testFeaturesFromList(uint256 tokenId) public {
        tokenId = bound(tokenId, 0, 9999);
        assertEq(manager.getFeatures(tokenId), getFeatures(tokenId));
    }
}
