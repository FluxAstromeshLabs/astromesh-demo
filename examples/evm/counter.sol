// SPDX-License-Identifier: MIT
pragma solidity >=0.7.0 <0.9.0;

contract Storage {
    uint256 counter;

    function add(uint256 amount) public {
        counter += amount;
    }

    function get() public view returns (uint256){
        return counter;
    }
}