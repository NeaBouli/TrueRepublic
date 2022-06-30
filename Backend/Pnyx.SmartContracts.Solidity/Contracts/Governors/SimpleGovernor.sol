// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/governance/Governor.sol";

contract SimpleGovernor is Governor {
    constructor(string memory name_) Governor(name_) {
    }
}