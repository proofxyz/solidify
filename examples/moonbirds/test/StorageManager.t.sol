// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "forge-std/Test.sol";
import "forge-std/console2.sol";
import "./TestLib.sol";

import {AssetStorageManager} from "moonbirds-inchain/AssetStorageManager.sol";
import {LayerStorageDeployer} from
    "moonbirds-inchain/gen/LayerStorageDeployer.sol";
import {TraitStorageDeployer} from
    "moonbirds-inchain/gen/TraitStorageDeployer.sol";

import {TraitType} from "moonbirds-inchain/gen/TraitStorageMapping.sol";

contract StorageManagerTest is Test {
    using TestLib for Vm;

    AssetStorageManager public manager;

    function setUp() public {
        manager = new AssetStorageManager(
            LayerStorageDeployer.deployAsStatic(),
            TraitStorageDeployer.deployAsStatic()
        );
    }

    function testTraitsFromList() public {
        assertEq(
            manager.loadTrait(TraitType.Background, 8), "Enlightened Purple"
        );
    }
}
