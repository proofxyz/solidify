// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity 0.8.16;

import {IBucketStorage} from "solidify-contracts/IBucketStorage.sol";
import {Compressed} from "solidify-contracts/Compressed.sol";
import {InflateLibWrapper} from "solidify-contracts/InflateLibWrapper.sol";
import {LabelledBucketLib} from "solidify-contracts/LabelledBucketLib.sol";
import {BucketCoordinates} from "solidify-contracts/BucketStorageLib.sol";

import {Features, FeaturesLib} from "moonbirds-inchain/gen/Features.sol";
import {FeaturesStorageMapping} from
    "moonbirds-inchain/gen/FeaturesStorageMapping.sol";
import {FeaturesStorageDeployer} from
    "moonbirds-inchain/gen/FeaturesStorageDeployer.sol";

/**
 * @notice Keeps records of all deployed BucketStorages that contain Moobird
 * features and provides an abstraction layer that allows features to be
 * accessed via tokenIDs.
 */
contract FeaturesStorageManager {
    using LabelledBucketLib for bytes;
    using FeaturesLib for Features;
    using FeaturesLib for bytes;
    using InflateLibWrapper for Compressed;

    // =========================================================================
    //                           Storage
    // =========================================================================

    /**
     * @notice Bundle of `BucketStorage`s containing moonbird features.
     */
    FeaturesStorageDeployer.Bundle private _bundle;

    // =========================================================================
    //                           Constructor
    // =========================================================================

    constructor(FeaturesStorageDeployer.Bundle memory bundle_) {
        _bundle = bundle_;
    }

    function getFeatures(uint256 tokenId)
        external
        view
        returns (Features memory)
    {
        BucketCoordinates memory bucket = FeaturesStorageMapping.locate(tokenId);

        return _bundle.storages[bucket.storageId].getBucket(bucket.bucketId)
            .inflate().findFieldByLabel(uint16(tokenId), 7).deserialise();
    }
}
