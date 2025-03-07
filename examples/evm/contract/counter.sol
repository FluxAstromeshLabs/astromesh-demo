// SPDX-License-Identifier: MIT
pragma solidity >=0.7.0 <0.9.0;

contract Storage {
    uint256 counter;

    function count() public {
        counter += 1;
    }

    function get() public view returns (uint256){
        return counter;
    }
}