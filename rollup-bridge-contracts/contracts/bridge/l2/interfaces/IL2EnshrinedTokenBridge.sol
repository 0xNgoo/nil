// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import { IL2Bridge } from "./IL2Bridge.sol";

interface IL2EnshrinedTokenBridge is IL2Bridge {
  /*//////////////////////////////////////////////////////////////////////////
                             ERRORS   
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Thrown when the L1 token address is invalid
  error ErrorInvalidL1TokenAddress();

  /// @notice Thrown when the token address is invalid
  error ErrorInvalidTokenAddress();

  /// @notice Thrown when the L1 token address does not match the expected address
  error ErrorL1TokenAddressMismatch();

  /*//////////////////////////////////////////////////////////////////////////
                             EVENTS   
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Emitted when the token mapping is updated
  /// @param l2EnshrinedTokenAddress The address of the enshrined token on L2
  /// @param l1TokenAddress The address of the corresponding token on L1
  event TokenMappingUpdated(address indexed l2EnshrinedTokenAddress, address indexed l1TokenAddress);

  /// @notice Emitted when ERC20 token is deposited from L1 to L2 and transfer to recipient.
  /// @param l1Token The address of the token in L1.
  /// @param l2Token The address of the token in L2.
  /// @param from The address of sender in L1.
  /// @param to The address of recipient in L2.
  /// @param feeRefundRecipient The address of excess-fee refund recipient on L2.
  /// @param amount The amount of token withdrawn from L1 to L2.
  /// @param data The optional calldata passed to recipient in L2.
  event FinalizeDepositERC20(
    address indexed l1Token,
    address indexed l2Token,
    address indexed from,
    address to,
    address feeRefundRecipient,
    uint256 amount,
    bytes data
  );

  /*//////////////////////////////////////////////////////////////////////////
                             PUBLIC CONSTANT FUNCTIONS   
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Return the corresponding l1 token address given l2 token address.
  /// @param l2Token The address of l2 token.
  function getL1ERC20Address(address l2Token) external view returns (address);

  /*//////////////////////////////////////////////////////////////////////////
                             PUBLIC MUTATION FUNCTIONS   
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Complete a deposit from L1 to L2 and send fund to recipient's account in L2.
  /// @dev Make this function payable to handle WETH deposit/withdraw.
  ///      The function should only be called by L2ScrollMessenger.
  ///      The function should also only be called by L1ERC20Gateway in L1.
  /// @param l1Token The address of corresponding L1 token.
  /// @param l2Token The address of corresponding L2 token.
  /// @param depositor The address of account who deposits the token in L1.
  /// @param recipient The address of recipient in L2 to receive the token.
  /// @param feeRefundRecipient The address of excess-fee refund recipient on L2.
  /// @param depositAmount The amount of the token to deposit.
  /// @param targetCallData Optional data to forward to recipient's account.
  function finalizeERC20Deposit(
    address l1Token,
    address l2Token,
    address depositor,
    address recipient,
    address feeRefundRecipient,
    uint256 depositAmount,
    bytes calldata targetCallData
  ) external payable;

  /// @notice Sets the token mapping between L2 enshrined token and L1 token
  /// @param l2EnshrinedTokenAddress The address of the enshrined token on L2
  /// @param l1TokenAddress The address of the corresponding token on L1
  function setTokenMapping(address l2EnshrinedTokenAddress, address l1TokenAddress) external;

  /**
   * @notice Pauses or unpauses the contract.
   * @dev This function allows the owner to pause or unpause the contract.
   * @param status The pause status to update.
   */
  function setPause(bool status) external;

  /**
   * @notice transfers ownership to the newOwner.
   * @dev This function revokes the `OWNER_ROLE` from the current owner, calls `acceptOwnership` using
   * OwnableUpgradeable's `transferOwnership` transfer the owner rights to newOwner
   * @param newOwner The address of the new owner.
   */
  function transferOwnershipRole(address newOwner) external;
}
