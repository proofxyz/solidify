// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity 0.8.16;

/**
 * @notice Key (name) / value pair for attributes, e.g. 'Body: Professor'
 */
struct Attribute {
    string name;
    string value;
}

/**
 * @notice Utility library to work with multiple attributes.
 */
library AttributesLib {
    /**
     * @notice Thrown if the attribute buffer capacity is exceeded.
     */
    error BufferOverflow();

    /**
     * @notice The capacity of new attribute buffers.
     * @dev We set this to a constant value of 8 because all Moonbirds have max.
     * 8 attributes.
     */
    uint256 internal constant _BUFFER_CAPACITY = 8;

    /**
     * @notice Allocates a new attributes buffer to be appended to.
     */
    function newBuffer() internal pure returns (Attribute[] memory) {
        Attribute[] memory buffer = new Attribute[](_BUFFER_CAPACITY);
        assembly {
            mstore(buffer, 0)
        }
        return buffer;
    }

    /**
     * @notice Adds an attribute to the buffer.
     * @dev Reverts if the buffers capacity is exceeded.
     */
    function addAttribute(
        Attribute[] memory buffer,
        string memory name,
        string memory value
    ) internal pure {
        if (bytes(value).length == 0) {
            return;
        }

        uint256 len = buffer.length;
        if (len == _BUFFER_CAPACITY) {
            revert BufferOverflow();
        }

        assembly {
            mstore(buffer, add(len, 1))
        }

        buffer[len] = Attribute({name: name, value: value});
    }

    /**
     * @notice Convenience overload of `addAttribute` assuming that the value
     * bytes encode a string.
     * @dev See above.
     */
    function addAttribute(
        Attribute[] memory buffer,
        string memory name,
        bytes memory value
    ) internal pure {
        addAttribute(buffer, name, string(value));
    }
}
