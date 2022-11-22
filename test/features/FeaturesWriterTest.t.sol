// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "forge-std/console2.sol";
import "forge-std/Test.sol";

import {MerkleProof} from
    "openzeppelin-contracts/utils/cryptography/MerkleProof.sol";

import {IBucketStorage} from "solidify-contracts/IBucketStorage.sol";
import {
    BucketStorageLib,
    BucketCoordinates
} from "solidify-contracts/BucketStorageLib.sol";
import {LabelledBucketLib} from "solidify-contracts/LabelledBucketLib.sol";

import {Features, FeatureType, FeaturesLib} from "./gen/Features.sol";
import {FeaturesStorageDeployer} from "./gen/FeaturesStorageDeployer.sol";
import {FeaturesStorageMapping} from "./gen/FeaturesStorageMapping.sol";
import {FeaturesLoader} from "./FeaturesLoader.sol";

contract FeaturesWriterTest is Test {
    using BucketStorageLib for IBucketStorage[];
    using LabelledBucketLib for bytes;
    using FeaturesLib for Features;
    using FeaturesLib for bytes;

    IBucketStorage[] public bundle;
    FeaturesLoader public loader;

    constructor() {
        bundle = FeaturesStorageDeployer.deployAsDynamic();
        loader = new FeaturesLoader();
    }

    function testBundleMetadata() public {
        assertEq(bundle.length, 2);

        assertEq(bundle[0].numBuckets(), 2);
        assertEq(bundle[0].numFields(), 4);
        assertEq(bundle[0].numFieldsPerBucket()[0], 2);
        assertEq(bundle[0].numFieldsPerBucket()[1], 2);

        assertEq(bundle[1].numBuckets(), 1);
        assertEq(bundle[1].numFields(), 1);
        assertEq(bundle[1].numFieldsPerBucket()[0], 1);
    }

    function _loadLabelled(uint256 storageId, uint256 bucketId, uint16 label)
        internal
        view
        returns (Features memory)
    {
        BucketCoordinates memory bucket =
            BucketCoordinates({storageId: storageId, bucketId: bucketId});
        return bundle.loadUncompressed(bucket).findFieldByLabel(
            label, FeaturesLib.FEATURES_LENGTH
        ).deserialise();
    }

    function _loadMapped(uint16 tokenId)
        internal
        view
        returns (Features memory)
    {
        BucketCoordinates memory bucket = FeaturesStorageMapping.locate(tokenId);

        return bundle.loadUncompressed(bucket).findFieldByLabel(
            tokenId, FeaturesLib.FEATURES_LENGTH
        ).deserialise();
    }

    function testFieldAccess() public {
        assertEq(_loadLabelled(0, 0, 0), Features({foo: 0, bar: 1, qux: 1}));
        assertEq(_loadLabelled(0, 0, 1), Features({foo: 2, bar: 3, qux: 0}));
        assertEq(_loadLabelled(0, 1, 2), Features({foo: 1, bar: 2, qux: 0}));
        assertEq(_loadLabelled(0, 1, 6), Features({foo: 0, bar: 2, qux: 0}));
        assertEq(_loadLabelled(1, 0, 7), Features({foo: 0, bar: 3, qux: 1}));
    }

    function testMapping() public {
        assertEq(_loadMapped(0), Features({foo: 0, bar: 1, qux: 1}));
        assertEq(_loadMapped(1), Features({foo: 2, bar: 3, qux: 0}));
        assertEq(_loadMapped(2), Features({foo: 1, bar: 2, qux: 0}));
        assertEq(_loadMapped(6), Features({foo: 0, bar: 2, qux: 0}));
        assertEq(_loadMapped(7), Features({foo: 0, bar: 3, qux: 1}));
    }

    function testDebugJson() public {
        assertEq(loader.getFeatures(0), Features({foo: 0, bar: 1, qux: 1}));
        assertEq(loader.getFeatures(1), Features({foo: 2, bar: 3, qux: 0}));
        assertEq(loader.getFeatures(2), Features({foo: 1, bar: 2, qux: 0}));
        assertEq(loader.getFeatures(3), Features({foo: 1, bar: 2, qux: 1}));
        assertEq(loader.getFeatures(4), Features({foo: 1, bar: 1, qux: 1}));
        assertEq(loader.getFeatures(5), Features({foo: 1, bar: 0, qux: 0}));
        assertEq(loader.getFeatures(6), Features({foo: 0, bar: 2, qux: 0}));
        assertEq(loader.getFeatures(7), Features({foo: 0, bar: 3, qux: 1}));
    }

    function testMerkle() public {
        assertEq(loader.getMerkleRoot(), FeaturesLib.FEATURES_ROOT);

        for (uint256 i; i < 8; ++i) {
            assertTrue(
                MerkleProof.verify(
                    loader.getProof(i),
                    FeaturesLib.FEATURES_ROOT,
                    loader.getFeatures(i).hash(i)
                )
            );
        }
    }

    function testWrongMerkleProof() public {
        assertFalse(
            MerkleProof.verify(
                loader.getProof(0),
                FeaturesLib.FEATURES_ROOT,
                loader.getFeatures(1).hash(1)
            )
        );
    }

    function assertEq(Features memory got, Features memory want) public {
        assertEq(FeaturesLib.serialise(got), FeaturesLib.serialise(want));
    }
}
