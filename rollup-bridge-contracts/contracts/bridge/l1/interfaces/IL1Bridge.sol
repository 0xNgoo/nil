// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import { IBridge } from "../../interfaces/IBridge.sol";

interface IL1Bridge is IBridge {
    error UnAuthorizedCaller();
    error InvalidDepositType();

    /*//////////////////////////////////////////////////////////////////////////
                             PUBLIC MUTATION FUNCTIONS   
    //////////////////////////////////////////////////////////////////////////*/

    function cancelDeposit(bytes32 messageHash) external payable;

    /**
     * @notice Pauses or unpauses the contract.
     * @dev This function allows the owner to pause or unpause the contract.
     * @param _status The pause status to update.
     */
    function setPause(bool _status) external;

    function acceptOwnership() external;

    /*//////////////////////////////////////////////////////////////////////////
                             PUBLIC CONSTANT FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /// @notice The address of L1BridgeRouter/L2BridgeRouter contract.
    function router() external view returns (address);

    /// @notice The address of Bridge contract on other side (for L1Bridge it would be the bridge-address on L2 and for
    /// L2Bridge this would be the bridge-address on L1)
    function counterpartyBridge() external view returns (address);

    /// @notice The address of corresponding L1NilMessenger/L2NilMessenger contract.
    function messenger() external view returns (address);
}
