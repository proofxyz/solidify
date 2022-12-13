// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "forge-std/console2.sol";
import "forge-std/Test.sol";

import {GroupStorageStorageDeployer} from
    "./gen/GroupStorageStorageDeployer.sol";
import {
    GroupStorageType,
    GroupStorageStorageMapping
} from "./gen/GroupStorageStorageMapping.sol";

import {IBucketStorage} from "solidify-contracts/IBucketStorage.sol";
import {
    BucketStorageLib,
    BucketCoordinates,
    FieldCoordinates
} from "solidify-contracts/BucketStorageLib.sol";
import {IndexedBucketLib} from "solidify-contracts/IndexedBucketLib.sol";

contract IndexedBucketsTest is Test {
    using BucketStorageLib for IBucketStorage[];
    using IndexedBucketLib for bytes;

    IBucketStorage[] public bundle;

    constructor() {
        bundle = GroupStorageStorageDeployer.deployAsDynamic();
    }

    function testBundleMetadata() public {
        assertEq(bundle.length, 2);

        assertEq(bundle[0].numBuckets(), 2);
        assertEq(bundle[0].numFields(), 5);
        assertEq(bundle[0].numFieldsPerBucket()[0], 2);
        assertEq(bundle[0].numFieldsPerBucket()[1], 3);

        assertEq(bundle[1].numBuckets(), 1);
        assertEq(bundle[1].numFields(), 1);
        assertEq(bundle[1].numFieldsPerBucket()[0], 1);
    }

    function _loadIndexed(uint256 storageId, uint256 bucketId, uint256 fieldId)
        internal
        view
        returns (string memory)
    {
        BucketCoordinates memory bucket =
            BucketCoordinates({storageId: storageId, bucketId: bucketId});
        return string(bundle.loadUncompressed(bucket).getField(fieldId));
    }

    function _loadMapped(GroupStorageType typ, uint256 index)
        internal
        view
        returns (string memory)
    {
        GroupStorageStorageMapping.StorageCoordinates memory coords =
            GroupStorageStorageMapping.locate(typ, index);

        return string(
            bundle.loadUncompressed(coords.bucket).getField(coords.fieldId)
        );
    }

    function testFieldAccess() public {
        assertEq(_loadIndexed(0, 0, 0), "foo0");
        assertEq(_loadIndexed(0, 0, 1), "foo1");
        assertEq(_loadIndexed(0, 1, 0), "bar0");
        assertEq(_loadIndexed(0, 1, 1), "bar1");
        assertEq(_loadIndexed(0, 1, 2), "bar2");
        assertEq(_loadIndexed(1, 0, 0), "qux0");
    }

    function testMapping() public {
        assertEq(_loadMapped(GroupStorageType.FOO, 0), "foo0");
        assertEq(_loadMapped(GroupStorageType.FOO, 1), "foo1");
        assertEq(_loadMapped(GroupStorageType.BAR, 0), "bar0");
        assertEq(_loadMapped(GroupStorageType.BAR, 1), "bar1");
        assertEq(_loadMapped(GroupStorageType.BAR, 2), "bar2");
        assertEq(_loadMapped(GroupStorageType.QUX, 0), "qux0");
    }

    function testLocateByFieldGroups() public {
        _testLocateByFieldGroupAndIndex(0, 0, 0, 0, 0);
        _testLocateByFieldGroupAndIndex(0, 1, 0, 0, 1);
        _testLocateByFieldGroupAndIndex(1, 0, 0, 1, 0);
        _testLocateByFieldGroupAndIndex(1, 1, 0, 1, 1);
        _testLocateByFieldGroupAndIndex(1, 2, 0, 1, 2);
        _testLocateByFieldGroupAndIndex(2, 0, 1, 0, 0);
    }

    function _testLocateByFieldGroupAndIndex(
        uint256 fieldGroupId,
        uint256 fieldIdx,
        uint256 wantStorageId,
        uint256 wantBucketId,
        uint256 wantFieldId
    ) public {
        uint256[] memory fieldGroupSizes = new uint256[](3);
        fieldGroupSizes[0] = 2;
        fieldGroupSizes[1] = 3;
        fieldGroupSizes[2] = 1;

        FieldCoordinates memory coords = bundle.locateByFieldGroupAndIndex(
            fieldGroupSizes, fieldGroupId, fieldIdx
        );

        assertEq(coords.bucket.storageId, wantStorageId);
        assertEq(coords.bucket.bucketId, wantBucketId);
        assertEq(coords.fieldId, wantFieldId);
    }
}
