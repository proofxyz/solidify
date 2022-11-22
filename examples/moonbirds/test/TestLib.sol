// SPDX-License-Identifier: MIT
// Copyright 2022 PROOF Holdings Inc
pragma solidity ^0.8.15;

import {Vm, Test, stdJson} from "forge-std/Test.sol";
import {console2} from "forge-std/console2.sol";
import {Attribute} from "moonbirds-inchain/types/Attribute.sol";
import {Features, FeaturesLib} from "moonbirds-inchain/gen/Features.sol";

library TestLib {
    using stdJson for string;

    function writeFile(Vm vm, bytes memory data, string memory filename)
        public
    {
        vm.writeFile(filename, vm.toString(data));
        string[] memory cmds = new string[](2);
        cmds[0] = "./bin/convertHexFile";
        cmds[1] = filename;
        vm.ffi(cmds);
    }

    function mktemp(Vm vm, string memory suffix)
        public
        returns (string memory)
    {
        string[] memory cmds = new string[](2);
        cmds[0] = "mktemp";
        cmds[1] = string.concat("--suffix=", suffix);
        string memory filename = string(vm.ffi(cmds));
        return filename;
    }

    function writeTempFile(Vm vm, bytes memory data)
        public
        returns (string memory)
    {
        return writeTempFile(vm, data, "");
    }

    function writeTempFile(Vm vm, bytes memory data, string memory suffix)
        public
        returns (string memory)
    {
        string memory tmp = mktemp(vm, suffix);
        writeFile(vm, data, tmp);
        return tmp;
    }

    function isValidBMP(Vm vm, bytes memory bmp) public returns (bool) {
        string memory filename = writeTempFile(vm, bmp, "");
        string[] memory cmds = new string[](2);
        cmds[0] = "./bin/isValidBMP";
        cmds[1] = filename;
        bytes memory re = vm.ffi(cmds);

        // Checking the echoed exit code of the script
        return keccak256(re) == keccak256("0");
    }

    function matchesReferenceArtwork(Vm vm, bytes memory bmp, uint256 birbId)
        public
        returns (bool)
    {
        string memory filename = writeTempFile(vm, bmp, vm.toString(birbId));

        string[] memory cmds = new string[](3);
        cmds[0] = "./bin/isSameImage";
        cmds[1] = filename;
        cmds[2] = string(
            string.concat(
                "./assets/moonbirds-assets/collection/png/",
                vm.toString(birbId),
                ".png"
            )
        );
        bytes memory re = vm.ffi(cmds);

        // Checking the echoed exit code of the script
        return keccak256(re) == keccak256("0");
    }

    function isCorrectBirbAttributes(
        Vm vm,
        uint256 birbId,
        Attribute[] memory attributes
    ) public returns (bool) {
        string memory filename = mktemp(vm, vm.toString(birbId));

        for (uint256 i; i < attributes.length; ++i) {
            vm.writeLine(
                filename,
                string.concat(attributes[i].name, ",", attributes[i].value)
            );
        }

        uint256 cursor = 0;
        string[] memory cmds = new string[](13);
        cmds[cursor++] = "./bin/validateAttributes";

        cmds[cursor++] = "--tokenID";
        cmds[cursor++] = vm.toString(birbId);

        cmds[cursor++] = "--testAttributesCSVPath";
        cmds[cursor++] = filename;

        cmds[cursor++] = "--refMetadataPath";
        cmds[cursor++] = "./assets/reference-metadata.json";

        cmds[cursor++] = "--ignoreRefAttribute";
        cmds[cursor++] = "ID";

        cmds[cursor++] = "--ignoreRefAttribute";
        cmds[cursor++] = "Eye Color";

        cmds[cursor++] = "--ignoreRefAttribute";
        cmds[cursor++] = "Beak Color";

        bytes memory re = vm.ffi(cmds);

        // Checking the echoed exit code of the script
        return keccak256(re) == keccak256("0");
    }
}

contract MBTest is Test {
    using TestLib for Vm;

    function assertEq(Features memory got, Features memory want) public {
        assertEq(FeaturesLib.serialise(got), FeaturesLib.serialise(want));
    }

    function assertNeq(string memory got, string memory notWant) public {
        assertFalse(
            keccak256(bytes(got)) == keccak256(bytes(notWant)),
            "Strings are equal"
        );
    }
}
