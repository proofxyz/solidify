// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity >=0.8.16 <0.9.0;

import {IBucketStorage} from "solidify-contracts/IBucketStorage.sol";
import {
    InflateLibWrapper,
    Compressed
} from "solidify-contracts/InflateLibWrapper.sol";
import {IndexedBucketLib} from "solidify-contracts/IndexedBucketLib.sol";
import {LabelledBucketLib} from "solidify-contracts/LabelledBucketLib.sol";

/**
 * @notice Coordinates to identify a bucket inside a storage bundle.
 * @dev These describe a hierarchical storage structure akin to
 * `x.storageId.bucketId`
 */
struct BucketCoordinates {
    uint256 storageId;
    uint256 bucketId;
}

/**
 * @notice Coordinates to identify a field inside a storage bundle.
 * @dev These describe a hierarchical storage structure akin to
 * `x.storageId.bucketId.fieldId`
 */
struct FieldCoordinates {
    BucketCoordinates bucket;
    uint256 fieldId;
}

/**
 * @notice Utility library to retrieve data from a storage bundle.
 */
library BucketStorageLib {
    using InflateLibWrapper for Compressed;

    /**
     * @notice Retrieves uncompressed bucket data from a bundle.
     */
    function loadUncompressed(
        IBucketStorage[] storage bundle,
        BucketCoordinates memory coordinates
    ) internal view returns (bytes memory) {
        return loadCompressed(bundle, coordinates).inflate();
    }

    /**
     * @notice Retrieves compressed bucket data from a bundle.
     */
    function loadCompressed(
        IBucketStorage[] storage bundle,
        BucketCoordinates memory coordinates
    ) internal view returns (Compressed memory) {
        return bundle[coordinates.storageId].getBucket(coordinates.bucketId);
    }

    /**
     * @notice Computes the total number of fields in a bucket storage bundle.
     */
    function numFields(IBucketStorage[] storage bundle)
        internal
        view
        returns (uint256)
    {
        uint256 numFields_;
        uint256 len = bundle.length;
        for (uint256 storageId; storageId < len; ++storageId) {
            numFields_ += bundle[storageId].numFields();
        }
        return numFields_;
    }

    /**
     * @notice Computes the storage coordinates of a field in a bundle
     * identified by a (group, index) pair.
     * @dev This is analogous to the sequential StorageMappings that `solidify`
     * produces, but applied to a dynamic bundle instead.
     * @param bundle The bundle of bucket storages
     * @param fieldGroupSizes the number of fields in each FieldGroup
     * @param fieldGroupId the id of the group of interest (matching the sizes
     * given above)
     * @param index the index of the field of interest in `fieldGroupId`
     */
    function locateByFieldGroupAndIndex(
        IBucketStorage[] storage bundle,
        uint256[] memory fieldGroupSizes,
        uint256 fieldGroupId,
        uint256 index
    ) internal view returns (FieldCoordinates memory) {
        uint256 globalFieldId = _computeAbsoluteFieldIdFromGroups(
            fieldGroupSizes, fieldGroupId, index
        );
        return locateByAbsoluteFieldId(bundle, globalFieldId);
    }

    /**
     * @notice Retrieves the storage coordinates of a field identified by its
     * global `fieldId`.
     * @dev The association between global fieldId and stored field is
     * established by hierarchically iterating through each structure.
     * We start our count at the first Bucket of the first BucketStorage. The
     * fields therin will have indices `0..bundle[0].numFieldsPerBucket()[0]`.
     * Then we continue with the second Bucket in the same Storage, and so
     * on. Once we have exhausted all the Buckets in the first Storage, we
     * move on to the next Storage - again starting at the first Bucket.
     */
    function locateByAbsoluteFieldId(
        IBucketStorage[] storage bundle,
        uint256 absoluteFieldId
    ) internal view returns (FieldCoordinates memory) {
        uint256 storageId;
        uint256 len = bundle.length;
        uint256 fieldId;

        for (; storageId < len; ++storageId) {
            uint256 numFields_ = bundle[storageId].numFields();

            if (fieldId + numFields_ > absoluteFieldId) {
                break;
            }
            fieldId += numFields_;
        }

        uint256[] memory numFieldsPerBucket =
            bundle[storageId].numFieldsPerBucket();

        uint256 bucketId;
        for (; bucketId < numFieldsPerBucket.length; ++bucketId) {
            if (fieldId + numFieldsPerBucket[bucketId] > absoluteFieldId) {
                break;
            }
            fieldId += numFieldsPerBucket[bucketId];
        }

        return FieldCoordinates({
            bucket: BucketCoordinates({storageId: storageId, bucketId: bucketId}),
            fieldId: absoluteFieldId - fieldId
        });
    }

    /**
     * @notice Computes the absolute field index of a (group, index) pair.
     * @dev The absolute field index is computed by sequentially iterating
     * through the groups defined by `fieldGroupSizes`
     * @param fieldGroupSizes the number of fields in each FieldGroup
     * @param fieldGroupId the id of the group of interest (matching the sizes
     * given above)
     * @param index the index of the field of interest in `fieldGroupId`
     */
    function _computeAbsoluteFieldIdFromGroups(
        uint256[] memory fieldGroupSizes,
        uint256 fieldGroupId,
        uint256 index
    ) private pure returns (uint256) {
        uint256 fieldId;

        for (uint256 idx; idx < fieldGroupId; ++idx) {
            fieldId += fieldGroupSizes[idx];
        }
        fieldId += index;

        return fieldId;
    }
}
