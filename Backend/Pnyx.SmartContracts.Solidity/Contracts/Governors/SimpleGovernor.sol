// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/governance/Governor.sol";

contract SimpleGovernor is Governor {
    constructor(string memory name_) Governor(name_) {
    }

    function quorum(uint256 blockNumber) public view override returns (uint256) {
        return 0;
    }

    function _quorumReached(uint256 proposalId) internal view override returns (bool) {
        return false;
    }

    function _voteSucceeded(uint256 proposalId) internal view override returns (bool) {
        return false;
    }

    function _countVote(
        uint256 proposalId,
        address account,
        uint8 support,
        uint256 weight,
        bytes memory params
    ) internal override {

    }

    function votingPeriod() public view override returns (uint256) {
        return 0;
    }

    function votingDelay() public view override returns (uint256) {
        return 0;
    }

    function _getVotes(
        address account,
        uint256 blockNumber,
        bytes memory params
    ) internal view override returns (uint256) {
        return 0;
    }

    function COUNTING_MODE() public pure override returns (string memory) {
        return "";
    }

    function hasVoted(uint256 proposalId, address account) public view override returns (bool) {
        return false;
    }
}