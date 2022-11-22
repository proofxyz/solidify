// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity 0.8.16;

import {Attribute, AttributesLib} from "moonbirds-inchain/types/Attribute.sol";
import {Mutators} from "moonbirds-inchain/types/Mutators.sol";

import {Image, Rectangle} from "ethier/utils/Image.sol";
import {RawData} from "ethier/utils/RawData.sol";

import {AssetStorageManager} from "moonbirds-inchain/AssetStorageManager.sol";

import {TraitType} from "moonbirds-inchain/gen/TraitStorageMapping.sol";
import {LayerType} from "moonbirds-inchain/gen/LayerStorageMapping.sol";
import {Features, FeaturesLib} from "moonbirds-inchain/gen/Features.sol";

/**
 * @notice The Moonbirds artwork and attributes assembler.
 * @dev Loads layers or traits from storage based on the provided features and
 * assembles them into the final artwork or list of attributes, respectively.
 */
contract Assembler {
    using RawData for bytes;
    using Image for bytes;
    using AttributesLib for Attribute[];
    using FeaturesLib for Features;

    // =========================================================================
    //                           Constants
    // =========================================================================

    /**
     * @notice The handler providing access to stored image layer + trait data.
     */
    AssetStorageManager public immutable assetStorageManger;

    /**
     * @notice The native resolution of Moonbird images (42x42).
     */
    uint32 internal constant _NATIVE_MB_RES = 42;

    // =========================================================================
    //                           Constructor
    // =========================================================================
    constructor(AssetStorageManager assetStorageManger_) {
        assetStorageManger = assetStorageManger_;
    }

    /**
     * @notice Assembles the Moonbird pixel data based on the provided features
     * and mutators.
     * @return Raw pixel data of the moonbird image in row-major, BGR encoding.
     */
    function assembleArtwork(Features memory f, Mutators memory mutators)
        public
        view
        returns (bytes memory)
    {
        f.validate();

        bytes memory canvas = _getBackground(f, mutators.useProofBackground);
        canvas = _addLayerIfPresent(canvas, LayerType.Body, f.body);
        canvas = _addLayerIfPresent(canvas, LayerType.Eyes, f.eyes);
        canvas = _addLayerIfPresent(canvas, LayerType.Beak, f.beak);
        canvas = _addLayerIfPresent(canvas, LayerType.Eyewear, f.eyewear);
        canvas = _addLayerIfPresent(canvas, LayerType.Headwear, f.headwear);
        canvas = _addLayerIfPresent(canvas, LayerType.Outerwear, f.outerwear);

        return canvas;
    }

    /**
     * @notice Assembles the Moonbird pixel data based on the provided features.
     */
    function assembleAttributes(Features memory f)
        public
        view
        returns (Attribute[] memory)
    {
        f.validate();

        Attribute[] memory buffer = AttributesLib.newBuffer();

        if (f.background > 0) {
            buffer.addAttribute(
                "Background",
                assetStorageManger.loadTrait(
                    TraitType.Background, f.background - 1
                )
            );
        }

        if (f.beak > 0) {
            buffer.addAttribute(
                "Beak", assetStorageManger.loadTrait(TraitType.Beak, f.beak - 1)
            );
        }

        if (f.body > 0) {
            bytes memory body =
                bytes(assetStorageManger.loadTrait(TraitType.Body, f.body - 1));

            // The feather attribute is stored with the body trait, e.g.
            // "Emperor - Pink". We need to split this for the body and feather
            // attributes.

            bytes memory feathers = body.clone();

            uint256 len = bytes(body).length;
            for (uint256 i; i < len; ++i) {
                if (body[i] == "-") {
                    assembly {
                        mstore(body, sub(i, 1))
                        feathers := add(feathers, add(i, 2))
                        mstore(feathers, sub(len, add(i, 2)))
                    }
                    break;
                }
            }

            buffer.addAttribute("Body", body);
            buffer.addAttribute("Feathers", feathers);
        }

        if (f.eyes > 0) {
            buffer.addAttribute(
                "Eyes", assetStorageManger.loadTrait(TraitType.Eyes, f.eyes - 1)
            );
        }

        if (f.eyewear > 0) {
            buffer.addAttribute(
                "Eyewear",
                assetStorageManger.loadTrait(TraitType.Eyewear, f.eyewear - 1)
            );
        }

        if (f.headwear > 0) {
            buffer.addAttribute(
                "Headwear",
                assetStorageManger.loadTrait(TraitType.Headwear, f.headwear - 1)
            );
        }

        if (f.outerwear > 0) {
            buffer.addAttribute(
                "Outerwear",
                assetStorageManger.loadTrait(
                    TraitType.Outerwear, f.outerwear - 1
                )
            );
        }

        return buffer;
    }

    // =========================================================================
    //                            Internals
    // =========================================================================

    /**
     * @notice Initializes a pixel buffer with background data.
     */
    function _getBackground(Features memory f, bool useProofBackground)
        internal
        view
        returns (bytes memory)
    {
        // Load the PROOF background
        if (useProofBackground) {
            // Ignore the alpha info since we know that it will be zero.
            (bytes memory bgrPixelsProof,) = assetStorageManger.loadLayer(
                LayerType.Special, 0
            ).popByteFront();

            // The layer rectangle information can be ignored for backgrounds
            // because the fill the whole frame.
            (bgrPixelsProof,) = bgrPixelsProof.popDWORDFront();

            return bgrPixelsProof;
        }

        // Fill with solid color
        if (f.background < 8) {
            bytes memory canvas = new bytes(42 * 42 * 3);
            uint24[8] memory fixedColours = [
                0x000000, // None
                0x99CEFF, // Blue
                0xFF0000, // Glitch Red
                0xCED4D9, // Gray
                0x95DBAD, // Green
                0xFCB5DB, // Pink
                0xABA3FF, // Purple
                0xF5CD71 // Yellow
            ];

            canvas.fill(fixedColours[f.background]);
            return canvas;
        }

        // Load background gradient
        // Ignore the alpha info. See above
        (bytes memory bgrPixels,) = assetStorageManger.loadLayer(
            LayerType.Gradients, f.background - 8
        ).popByteFront();

        // Ignoring the rectangle info again. See above.
        (bgrPixels,) = bgrPixels.popDWORDFront();

        return bgrPixels;
    }

    /**
     * @notice Loads a given layer and alpha-blends it with the pixel buffer.
     */
    function _addLayerIfPresent(
        bytes memory canvas,
        LayerType layerType,
        uint8 layerValue
    ) internal view returns (bytes memory) {
        if (layerValue == 0) {
            return canvas;
        }

        (bytes memory data, bytes1 info) = assetStorageManger.loadLayer(
            layerType, layerValue - 1
        ).popByteFront();

        (bytes memory abgrPixels, bytes4 rect_) = data.popDWORDFront();

        bool hasAlpha = uint8(info) > 0;
        if (!hasAlpha) {
            // The full canvas would be overwritten - hence we can just return
            // the new pixels instead.
            return abgrPixels;
        }

        Rectangle memory rect = Rectangle({
            xMin: uint8(bytes1(rect_)),
            yMin: uint8(bytes1(rect_ << 8)),
            xMax: uint8(bytes1(rect_ << 16)),
            yMax: uint8(bytes1(rect_ << 24))
        });

        canvas.alphaBlend(abgrPixels, _NATIVE_MB_RES, rect);
        return canvas;
    }
}
