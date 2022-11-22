// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity >=0.8.16 <0.9.0;

import {RawData} from "ethier/utils/RawData.sol";

/**
 * @notice Utility library to retrieve label-prefixed fields from decompressed
 * Buckets.
 * @dev This library assumes that all fields have fixed length and start with an
 * strictly monotonically increasing, big-endian, uint16 label.
 * | ... | uint16 label | payload | ... |
 */
library LabelledBucketLib {
    using RawData for bytes;

    /**
     * @notice Thrown if a label is not between the first and last label in a
     * bucket.
     */
    error InvalidBinarySearchBound(
        uint16 label, uint16 leftBound, uint16 rightBound
    );

    /**
     * @notice Throws if a label cannot be found in the given bucket.
     */
    error LabelNotFound(uint256 label);

    /**
     * @notice Thrown if the bucket size cannot be divided into fields of given
     * length.
     */
    error BucketAndFieldLengthMismatch();

    /**
     * @notice Retrieves the field with a given label.
     * @dev Reverts if the label cannot be found.
     * @dev Retrieves the payload data in-memory to avoid reallocations.
     * This implies that the buffer data cannot be reused.
     * Intended syntax: `data = data.findFieldByLabel(label, fieldLength)`.
     * @param data The decompressed bucket data.
     * @param label The label of the field that should be retrieved.
     * @param fieldLength Number of payload bytes in a field.
     */
    function findFieldByLabel(
        bytes memory data,
        uint16 label,
        uint256 fieldLength
    ) internal pure returns (bytes memory) {
        uint256 chunkLength = fieldLength + 2;
        if (data.length % chunkLength != 0) {
            revert BucketAndFieldLengthMismatch();
        }
        uint256 idx = _binarySearchLabelled16Field(data, label, chunkLength);
        return data.slice(idx * chunkLength + 2, fieldLength);
    }

    /**
     * @notice Retrieves the field with a given label using a binary search.
     * @dev See also `findFieldByLabel`.
     */
    function _binarySearchLabelled16Field(
        bytes memory data,
        uint16 label,
        uint256 chunkLength
    ) private pure returns (uint256) {
        uint256 ia = 0;
        uint256 ib = data.length / chunkLength - 1;

        uint16 a = data.getUint16(ia * chunkLength);
        if (a == label) {
            return ia;
        }

        uint16 b = data.getUint16(ib * chunkLength);
        if (b == label) {
            return ib;
        }

        if (label < a) {
            revert InvalidBinarySearchBound(label, a, b);
        }
        if (b < label) {
            revert InvalidBinarySearchBound(label, a, b);
        }

        while (true) {
            if (ib - ia < 2) {
                // We cannot subdivide any further
                break;
            }

            // Compute new midpoint
            uint256 im = (ia + ib) >> 1;
            uint16 m = data.getUint16(im * chunkLength);

            if (m == label) {
                // Success
                return im;
            }

            if (m < label) {
                // Use the midpoint as new lower bound
                ia = im;
                a = m;
            } else {
                // Use the midpoint as new upper bound
                ib = im;
                b = m;
            }
        }

        revert LabelNotFound(label);
    }
}
