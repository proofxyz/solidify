// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import "forge-std/Test.sol";
import "forge-std/console2.sol";

import "./TestLib.sol";
import "../script/FeaturesLoader.sol";
import "../script/DeployHelper.sol";

import "moonbirds-inchain/gen/Features.sol";
import "ethier/utils/BMP.sol";
import "moonbirds-inchain/types/Mutators.sol";

contract AssemblerTest is Test, FeaturesLoader {
    using TestLib for Vm;
    using FeaturesLib for Features;

    bool public constant DEBUG = true;

    Assembler public assembler;

    constructor() {
        DeployHelper deployer = new DeployHelper();
        deployer.deployAssembler();

        assembler = deployer.assembler();
    }

    function testBirb() public {
        uint256 iBirb = 10;
        bytes memory bmp = BMP.bmp(
            assembler.assembleArtwork(
                getFeatures(iBirb), Mutators({useProofBackground: false})
            ),
            42,
            42
        );
        assertTrue(vm.matchesReferenceArtwork(bmp, iBirb));
    }

    function assertMetadataEq(
        Attribute memory attr,
        string memory name,
        string memory value
    ) public {
        assertEq(
            keccak256(abi.encodePacked(attr.name)),
            keccak256(abi.encodePacked(name)),
            string.concat(attr.name, "!=", name)
        );
        assertEq(
            keccak256(abi.encodePacked(attr.value)),
            keccak256(abi.encodePacked(value)),
            string.concat(attr.value, "!=", value)
        );
    }

    function testMetadata() public {
        uint256 iBirb = 1;
        Attribute[] memory attr =
            assembler.assembleAttributes(getFeatures(iBirb));

        assertMetadataEq(attr[0], "Background", "Pink");
        assertMetadataEq(attr[1], "Beak", "Short");
        assertMetadataEq(attr[2], "Body", "Tabby");
        assertMetadataEq(attr[3], "Feathers", "Red");
        assertMetadataEq(attr[4], "Eyes", "Rainbow");
        assertMetadataEq(attr[5], "Outerwear", "Jean Jacket");
    }

    function _testBirb(uint256 iBirb, bool useProofBackground) internal {
        string memory fn = "/tmp/progress/";

        bytes memory bmp = BMP.bmp(
            assembler.assembleArtwork(
                getFeatures(iBirb),
                Mutators({useProofBackground: useProofBackground})
            ),
            42,
            42
        );
        bool successArtwork = vm.matchesReferenceArtwork(bmp, iBirb);

        Attribute[] memory attr =
            assembler.assembleAttributes(getFeatures(iBirb));
        bool successAttributes = vm.isCorrectBirbAttributes(iBirb, attr);

        assembly {
            successArtwork := xor(successArtwork, useProofBackground)
        }

        if (!successArtwork && DEBUG) {
            string memory path =
                vm.writeTempFile(bmp, string.concat(vm.toString(iBirb), ".bmp"));
            console2.log(path);
        }

        vm.writeLine(
            fn,
            string.concat(
                vm.toString(iBirb),
                " -> ",
                vm.toString(successArtwork),
                " ",
                vm.toString(successAttributes)
            )
        );
        assertTrue(successArtwork && successAttributes);
    }

    function _testBirbs(uint256 from, uint256 to) internal {
        string memory fn = "/tmp/progress/";
        vm.writeFile(fn, "");

        for (uint256 iBirb = from; iBirb < to; ++iBirb) {
            _testBirb(iBirb, false);
        }
    }

    function testSelection() public {
        uint16[15] memory selection = [
            940, // BG 1
            951, // BG 2
            1041, // BG 3
            1103, // BG 4
            1109, // BG 5
            1101, // BG 6
            3047, // Glitch
            1320, // Jade
            2080, // Enlightened
            1311, // Cosmic
            1787, // Robot
            2642, // Wonky Jade
            807,
            1759,
            1397
        ];
        for (uint256 i; i < selection.length; ++i) {
            _testBirb(selection[i], false);
        }
    }

    function testSelectionWithProofBG() public {
        uint16[15] memory selection = [
            940, // BG 1
            951, // BG 2
            1041, // BG 3
            1103, // BG 4
            1109, // BG 5
            1101, // BG 6
            3047, // Glitch
            1320, // Jade
            2080, // Enlightened
            1311, // Cosmic
            1787, // Robot
            2642, // Wonky Jade
            807,
            1759,
            1397
        ];
        for (uint256 i; i < selection.length; ++i) {
            _testBirb(selection[i], true);
        }
    }

    function testAllBirbs0() public {
        _testBirbs(0, 1000);
    }

    function testAllBirbs1() public {
        _testBirbs(1000, 2000);
    }

    function testAllBirbs2() public {
        _testBirbs(2000, 3000);
    }

    function testAllBirbs3() public {
        _testBirbs(3000, 4000);
    }

    function testAllBirbs4() public {
        _testBirbs(4000, 5000);
    }

    function testAllBirbs5() public {
        _testBirbs(5000, 6000);
    }

    function testAllBirbs6() public {
        _testBirbs(6000, 7000);
    }

    function testAllBirbs7() public {
        _testBirbs(7000, 8000);
    }

    function testAllBirbs8() public {
        _testBirbs(8000, 9000);
    }

    function testAllBirbs9() public {
        _testBirbs(9000, 10_000);
    }
}
