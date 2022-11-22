// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import "forge-std/Script.sol";
import "forge-std/console2.sol";

import {Vm, stdJson} from "forge-std/Components.sol";
import {Features, FeaturesLib} from "moonbirds-inchain/gen/Features.sol";

contract FeaturesLoader is Script {
    using stdJson for string;
    using FeaturesLib for Features;

    string public constant PATH = "./src/gen/features.json";

    Features[] private _features;
    bytes32[][] private _hashes;

    constructor() {
        _loadFeatures();
        _computeMerkleTree();
    }

    struct FeaturesJson {
        Features[] features;
    }

    function _loadFeatures() internal {
        string memory data = vm.readFile(PATH);
        bytes memory json = vm.parseJson(data);
        Features[] memory fs = abi.decode(json, (FeaturesJson)).features;

        for (uint256 i; i < fs.length; ++i) {
            _features.push(fs[i]);
        }
    }

    function getFeatures(uint256 tokenId)
        public
        view
        returns (Features memory)
    {
        return _features[tokenId];
    }

    function getProof(uint256 tokenId) public view returns (bytes32[] memory) {
        uint256 len = _hashes.length - 1;
        bytes32[] memory proof = new bytes32[](len);

        for (uint256 i; i < len; ++i) {
            bool odd = (tokenId % 2) == 1;
            uint256 neighbour = odd
                ? tokenId - 1
                : tokenId == _hashes[i].length - 1 ? tokenId : tokenId + 1;

            proof[i] = _hashes[i][neighbour];

            tokenId /= 2;
        }
        return proof;
    }

    function getMerkleRoot() public view returns (bytes32) {
        return _hashes[_hashes.length - 1][0];
    }

    function _computeMerkleTree() internal {
        _hashes.push(new bytes32[](_features.length));
        for (uint256 i; i < _features.length; ++i) {
            _hashes[0][i] = _features[i].hash(i);
        }

        for (uint256 i; true; ++i) {
            _hashes.push(_hashPair(_hashes[i]));
            if (_hashes[i + 1].length == 1) {
                break;
            }
        }
    }

    function _hashPair(bytes32[] memory leaves)
        internal
        pure
        returns (bytes32[] memory)
    {
        uint256 lenOld = leaves.length;
        uint256 lenNew = lenOld / 2;

        bool odd = lenOld % 2 != 0;

        bytes32[] memory h = new bytes32[](odd ? lenNew + 1 : lenNew);

        for (uint256 i; i < lenNew; ++i) {
            h[i] = _hashPair(leaves[2 * i], leaves[2 * i + 1]);
        }

        if (odd) {
            h[lenNew] = _hashPair(leaves[lenOld - 1], leaves[lenOld - 1]);
        }

        return h;
    }

    function _hashPair(bytes32 a, bytes32 b) private pure returns (bytes32) {
        return a < b ? _efficientHash(a, b) : _efficientHash(b, a);
    }

    function _efficientHash(bytes32 a, bytes32 b)
        private
        pure
        returns (bytes32 value)
    {
        /// @solidity memory-safe-assembly
        assembly {
            mstore(0x00, a)
            mstore(0x20, b)
            value := keccak256(0x00, 0x40)
        }
    }
}
