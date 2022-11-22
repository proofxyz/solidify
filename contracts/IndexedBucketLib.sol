// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity >=0.8.16 <0.9.0;

import {RawData} from "ethier/utils/RawData.sol";

/**
 * @notice Utility library to retrieve indexed fields from decompressed Buckets.
 * @dev This library assumes that the starting offsets of each fields are
 * stored sequentially as big-endian unt16 values at the start of the array
 * with the actual payload afterwards.
 * | uint16 offset field 0 | ... | uint16 offset field N-1 | payload 1 | ... |
 */
// Todo explain time/space complexity tradeoff of lenghtprefixed vs
// apriori-indexed
library IndexedBucketLib {
    using RawData for bytes;

    /**
     * @notice Thrown if a field index is not contained in a given bucket.
     */
    error FieldIndexOutOfBounds(uint256 fieldIndex, uint256 numFields);

    /**
     * @notice Retrieves the field with a given index.
     * @dev Retrieves the payload data in-memory to avoid reallocations.
     * This implies that the buffer data cannot be reused.
     * Intended syntax: `data = data.getField(idx)`.
     * @param data The decompressed bucket data.
     * @param fieldIdx The index of the field that should be retrieved.
     */
    function getField(bytes memory data, uint256 fieldIdx)
        internal
        pure
        returns (bytes memory)
    {
        // Since each index takes 2 bytes of storage, the number of fields can
        // be determined from the the location of the first field right after
        // the index header ends.
        uint256 numFields = data.getUint16(0) >> 1;
        if (fieldIdx >= numFields) {
            revert FieldIndexOutOfBounds(fieldIdx, numFields);
        }

        // The offset in the array at which the field of interest starts
        uint256 loc = data.getUint16(fieldIdx * 2);

        uint256 length;
        if (fieldIdx + 1 < numFields) {
            // The lenght of a field can be determined from the difference of
            // its starting offset to the one of the following field.
            length = data.getUint16((1 + fieldIdx) * 2) - loc;
        } else {
            // If the field is the last one in the array, we determine its end
            // from the full length of the array instead.
            length = data.length - loc;
        }

        // To save gas, we update the pointer and size in memory instead of
        // allocating new space and copying the content over.
        assembly {
            data := add(data, loc)
            mstore(data, length)
        }
        return data;
    }
}
