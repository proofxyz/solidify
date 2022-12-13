// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "forge-std/console2.sol";
import "forge-std/Test.sol";

import {
    IBucketStorage, Compressed
} from "solidify-contracts/IBucketStorage.sol";
import {
    BucketStorageLib,
    BucketCoordinates,
    FieldCoordinates
} from "solidify-contracts/BucketStorageLib.sol";
import {IndexedBucketLib} from "solidify-contracts/IndexedBucketLib.sol";

contract StubBucketStorage0 is IBucketStorage {
    function numBuckets() external pure returns (uint256) {
        return 2;
    }

    function numFields() external pure returns (uint256) {
        return 5;
    }

    function numFieldsPerBucket() external pure returns (uint256[] memory) {
        uint256[] memory ret = new uint[](2);
        ret[0] = 2;
        ret[1] = 3;
        return ret;
    }

    function getBucket(uint256 bucketIndex)
        external
        pure
        returns (Compressed memory)
    {}
}

contract StubBucketStorage1 is IBucketStorage {
    function numBuckets() external pure returns (uint256) {
        return 2;
    }

    function numFields() external pure returns (uint256) {
        return 6;
    }

    function numFieldsPerBucket() external pure returns (uint256[] memory) {
        uint256[] memory ret = new uint[](2);
        ret[0] = 1;
        ret[1] = 5;
        return ret;
    }

    function getBucket(uint256 bucketIndex)
        external
        pure
        returns (Compressed memory)
    {}
}

contract BucketStorageLibTest is Test {
    using BucketStorageLib for IBucketStorage[];

    IBucketStorage[] public bundle;

    constructor() {
        bundle.push(new StubBucketStorage0());
        bundle.push(new StubBucketStorage1());
    }

    function testLocateByAbsoluteFieldId() public {
        _testLocateByAbsoluteFieldId(0, 0, 0, 0);
        _testLocateByAbsoluteFieldId(1, 0, 0, 1);
        _testLocateByAbsoluteFieldId(2, 0, 1, 0);
        _testLocateByAbsoluteFieldId(3, 0, 1, 1);
        _testLocateByAbsoluteFieldId(4, 0, 1, 2);
        _testLocateByAbsoluteFieldId(5, 1, 0, 0);
        _testLocateByAbsoluteFieldId(6, 1, 1, 0);
        _testLocateByAbsoluteFieldId(7, 1, 1, 1);
        _testLocateByAbsoluteFieldId(8, 1, 1, 2);
        _testLocateByAbsoluteFieldId(9, 1, 1, 3);
        _testLocateByAbsoluteFieldId(10, 1, 1, 4);
    }

    function _testLocateByAbsoluteFieldId(
        uint256 absoluteFieldIdx,
        uint256 wantStorageId,
        uint256 wantBucketId,
        uint256 wantFieldId
    ) internal {
        FieldCoordinates memory coords =
            bundle.locateByAbsoluteFieldId(absoluteFieldIdx);

        assertEq(coords.bucket.storageId, wantStorageId, "wrong storage id");
        assertEq(coords.bucket.bucketId, wantBucketId, "wrong bucket id");
        assertEq(coords.fieldId, wantFieldId, "wrong field id");
    }

    function _testLocateByFieldGroupAndIndex(
        uint256 fieldGroupId,
        uint256 fieldIdx,
        uint256 wantStorageId,
        uint256 wantBucketId,
        uint256 wantFieldId
    ) internal {
        uint256[] memory fieldGroupSizes = new uint256[](4);
        fieldGroupSizes[0] = 3;
        fieldGroupSizes[1] = 5;
        fieldGroupSizes[2] = 2;
        fieldGroupSizes[3] = 1;

        FieldCoordinates memory coords = bundle.locateByFieldGroupAndIndex(
            fieldGroupSizes, fieldGroupId, fieldIdx
        );

        assertEq(coords.bucket.storageId, wantStorageId, "wrong storage id");
        assertEq(coords.bucket.bucketId, wantBucketId, "wrong bucket id");
        assertEq(coords.fieldId, wantFieldId, "wrong field id");
    }

    function testLocateByFieldGroups() public {
        _testLocateByFieldGroupAndIndex(0, 0, 0, 0, 0);
        _testLocateByFieldGroupAndIndex(0, 1, 0, 0, 1);
        _testLocateByFieldGroupAndIndex(0, 2, 0, 1, 0);
        _testLocateByFieldGroupAndIndex(1, 0, 0, 1, 1);
        _testLocateByFieldGroupAndIndex(1, 1, 0, 1, 2);
        _testLocateByFieldGroupAndIndex(1, 2, 1, 0, 0);
        _testLocateByFieldGroupAndIndex(1, 3, 1, 1, 0);
        _testLocateByFieldGroupAndIndex(1, 4, 1, 1, 1);
        _testLocateByFieldGroupAndIndex(2, 0, 1, 1, 2);
        _testLocateByFieldGroupAndIndex(2, 1, 1, 1, 3);
        _testLocateByFieldGroupAndIndex(3, 0, 1, 1, 4);
    }
}
