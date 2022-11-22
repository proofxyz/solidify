// SPDX-License-Identifier: MIT
pragma solidity 0.8.16;

import "forge-std/Script.sol";
import "forge-std/console2.sol";

import {LayerStorageDeployer} from
    "moonbirds-inchain/gen/LayerStorageDeployer.sol";
import {TraitStorageDeployer} from
    "moonbirds-inchain/gen/TraitStorageDeployer.sol";
import {AssetStorageManager} from "moonbirds-inchain/AssetStorageManager.sol";
import {Assembler} from "moonbirds-inchain/Assembler.sol";

import {FeaturesStorageDeployer} from
    "moonbirds-inchain/gen/FeaturesStorageDeployer.sol";
import {FeaturesStorageManager} from
    "moonbirds-inchain/FeaturesStorageManager.sol";

contract DeployHelper {
    AssetStorageManager public storageManager;
    Assembler public assembler;
    FeaturesStorageManager public proofFeaturesRegistry;

    function deployAssembler() public {
        if (address(storageManager) == address(0)) {
            storageManager = new AssetStorageManager(
                LayerStorageDeployer.deployAsStatic(),
                TraitStorageDeployer.deployAsStatic()
            );
        }
        if (address(assembler) == address(0)) {
            assembler = new Assembler(storageManager);
        }
    }

    function deployProofFeaturesRegistry() public {
        if (address(proofFeaturesRegistry) == address(0)) {
            proofFeaturesRegistry = new FeaturesStorageManager(
                FeaturesStorageDeployer.deployAsStatic()
            );
        }
    }
}
