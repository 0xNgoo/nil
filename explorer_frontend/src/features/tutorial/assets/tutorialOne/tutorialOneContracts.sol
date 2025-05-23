// SPDX-License-Identifier: MIT

pragma solidity ^0.8.9;

import "@nilfoundation/smart-contracts/contracts/Nil.sol";

// read more:
// https://docs.nil.foundation/nil/key-principles/async-execution
// https://docs.nil.foundation/nil/smart-contracts/handling-async-execution/

/**
 * @title Caller
 * @author =nil; Foundation
 * @notice The Caller contract must use sendValue() to send some
 * default tokens to the Receiver contract.
 * Caller must also be able to receive default tokens.
 */
contract Caller {
    using Nil for address;

    constructor() payable {}

    // Should send some default tokens to the Receiver contract
    // using Nil.asyncCall().
    function sendValue(address dst) public payable {
        // TODO: complete the function
    }
}

/**
 * @title Receiver
 * @author =nil; Foundation
 * @notice The Receiver contract must be able to receive default tokens
 * when the deposit() function is called.
 */
contract Receiver {
    // TODO: modify the function so that it can accept default tokens
    function deposit() public {}
}
