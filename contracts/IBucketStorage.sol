// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity >=0.8.16 <0.9.0;

import {Compressed} from "solidify-contracts/Compressed.sol";

/**
 * @notice BucketStorage is used to store a list of compressed buckets in
 * contract code.
 */
interface IBucketStorage {
    /**
     * @notice Thrown if a non-existant bucket should be accessed.
     */
    error InvalidBucketIndex();

    /**
     * @notice Returns the compressed bucket with given index.
     * @param bucketIndex The index of the bucket in the storage.
     * @dev Reverts if the index is out-of-range.
     */
    function getBucket(uint256 bucketIndex)
        external
        pure
        returns (Compressed memory);

    function numBuckets() external pure returns (uint256);

    function numFields() external pure returns (uint256);

    function numFieldsPerBucket() external pure returns (uint256[] memory);
}
