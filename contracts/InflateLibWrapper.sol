// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity >=0.8.16 <0.9.0;

import {InflateLib} from "inflate-sol/InflateLib.sol";
import {Compressed} from "solidify-contracts/Compressed.sol";

/**
 * @notice A lightweight convenience wrapper around `inflate-sol/InflateLib` to
 * make it compatible with our types.
 */
library InflateLibWrapper {
    /**
     * @notice Thrown on decompression errors.
     * @dev See `InflateLib.ErrorCode` for more details.
     */
    error InflationError(InflateLib.ErrorCode);

    /**
     * @notice Inflates compressed data.
     * @dev Reverts on decompression errors.
     */
    function inflate(Compressed memory data)
        internal
        pure
        returns (bytes memory)
    {
        (InflateLib.ErrorCode err, bytes memory inflated) =
            InflateLib.puff(data.data, data.uncompressedSize);

        if (err != InflateLib.ErrorCode.ERR_NONE) {
            revert InflationError(err);
        }

        return inflated;
    }
}

/**
 * @notice Public version of the above library to allow reuse through linking if
 * the performance overhead is not critical.
 */
library PublicInflateLibWrapper {
    using InflateLibWrapper for Compressed;

    function inflate(Compressed memory data)
        public
        pure
        returns (bytes memory)
    {
        return data.inflate();
    }
}
