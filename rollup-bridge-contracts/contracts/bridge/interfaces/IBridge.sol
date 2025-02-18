// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

interface IBridge {
    /*//////////////////////////////////////////////////////////////////////////
                             ERRORS
    //////////////////////////////////////////////////////////////////////////*/

    /// @dev Thrown when the given address is `address(0)`.
    error ErrorZeroAddress();
}
