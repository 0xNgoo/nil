// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

interface IL2WETHBridge {
  /*//////////////////////////////////////////////////////////////////////////
                             EVENTS   
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Emitted when ERC20 token is deposited from L1 to L2 and transfer to recipient.
  /// @param l1Token The address of the token in L1.
  /// @param l2Token The address of the token in L2.
  /// @param from The address of sender in L1.
  /// @param to The address of recipient in L2.
  /// @param feeRefundRecipient The address of excess-fee refund recipient on L2.
  /// @param amount The amount of token withdrawn from L1 to L2.
  /// @param data The optional calldata passed to recipient in L2.
  event FinalizedWETHDeposit(
    address indexed l1Token,
    address indexed l2Token,
    address indexed from,
    address to,
    address feeRefundRecipient,
    uint256 amount,
    bytes data
  );

  /*//////////////////////////////////////////////////////////////////////////
                            PUBLIC MUTATION FUNCTIONS      
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Complete a deposit from L1 to L2 and send fund to recipient's account in L2.
  /// @dev Make this function payable to handle WETH deposit/withdraw.
  ///      The function should only be called by L2ScrollMessenger.
  ///      The function should also only be called by L1ERC20Gateway in L1.
  /// @param l1Token The address of corresponding L1 token.
  /// @param l2Token The address of corresponding L2 token.
  /// @param from The address of account who deposits the token in L1.
  /// @param to The address of recipient in L2 to receive the token.
  /// @param feeRefundRecipient The address of excess-fee refund recipient on L2.
  /// @param amount The amount of the token to deposit.
  /// @param data Optional data to forward to recipient's account.
  function finalizeWETHDeposit(
    address l1Token,
    address l2Token,
    address from,
    address to,
    address feeRefundRecipient,
    uint256 amount,
    bytes calldata data
  ) external payable;
}
