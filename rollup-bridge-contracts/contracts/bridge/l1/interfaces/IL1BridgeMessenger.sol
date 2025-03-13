// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import { IBridgeMessenger } from "../../interfaces/IBridgeMessenger.sol";
import { INilGasPriceOracle } from "./INilGasPriceOracle.sol";

/// @title IL1BridgeMessenger
/// @notice Interface for the L1BridgeMessenger contract which handles cross-chain messaging between L1 and L2.
/// @dev This interface defines the functions and events for managing deposit messages, sending messages, and canceling
/// deposits.
interface IL1BridgeMessenger is IBridgeMessenger {
  /*//////////////////////////////////////////////////////////////////////////
                             ERRORS
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Thrown when a deposit message already exists.
  /// @param messageHash The hash of the deposit message.
  error DepositMessageAlreadyExist(bytes32 messageHash);

  /// @notice Thrown when a deposit message does not exist.
  /// @param messageHash The hash of the deposit message.
  error DepositMessageDoesNotExist(bytes32 messageHash);

  /// @notice Thrown when a deposit message is already cancelled.
  /// @param messageHash The hash of the deposit message.
  error DepositMessageAlreadyCancelled(bytes32 messageHash);

  /// @notice Thrown when a deposit message is not expired.
  /// @param messageHash The hash of the deposit message.
  error DepositMessageNotExpired(bytes32 messageHash);

  /// @notice Thrown when a message hash is not in the queue.
  /// @param messageHash The hash of the deposit message.
  error MessageHashNotInQueue(bytes32 messageHash);

  /// @notice Thrown when the max message processing time is invalid.
  error InvalidMaxMessageProcessingTime();

  /// @notice Thrown when the message cancel delta time is invalid.
  error InvalidMessageCancelDeltaTime();

  /// @notice Thrown when a bridge interface is invalid.
  error InvalidBridgeInterface();

  /// @notice Thrown when a bridge is already authorized.
  error BridgeAlreadyAuthorized();

  /// @notice Thrown when a bridge is not authorized.
  error BridgeNotAuthorized();

  /// @notice Thrown when any address other than l1NilRollup is attempting to remove messages from queue
  error NotAuthorizedToPopMessages();

  /// @notice Thrown when the deposit message has already been claimed.
  error DepositMessageAlreadyClaimed();

  /// @notice  Thrown when the deposit message hash is still in the queue, indicating that the message has not been executed on L2.
  error DepositMessageStillInQueue();

  /*//////////////////////////////////////////////////////////////////////////
                             EVENTS
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Emitted when a message is sent.
  /// @param messageSender The address of the message sender.
  /// @param messageTarget The address of the message recipient which can be an account/smartcontract.
  /// @param messageValue The amount of native eth sent to cover the l2-transaction fee and value associated if it is a
  /// native-eth deposit.
  /// @param messageNonce The nonce of the message.
  /// @param message The encoded message data.
  /// @param messageHash The hash of the message.
  /// @param depositType The type of the deposit.
  /// @param messageCreatedAt The time at which message was recorded.
  /// @param messageExpiryTime The expiry time of the message.
  /// @param l2FeeRefundAddress The recipient address for feeRefund on l2
  /// @param feeCreditData The feeCreditData struct with feeParameters snapshot from GasOracle and feeCredit captured from depositor
  event MessageSent(
    address indexed messageSender,
    address indexed messageTarget,
    uint256 messageValue,
    uint256 indexed messageNonce,
    bytes message,
    bytes32 messageHash,
    DepositType depositType,
    uint256 messageCreatedAt,
    uint256 messageExpiryTime,
    address l2FeeRefundAddress,
    INilGasPriceOracle.FeeCreditData feeCreditData
  );

  /// @notice Emitted when a deposit message is cancelled.
  /// @param messageHash The hash of the deposit message that was cancelled.
  event DepositMessageCancelled(bytes32 messageHash);

  /*//////////////////////////////////////////////////////////////////////////
                             MESSAGE STRUCTS   
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Enum representing the type of deposit.
  enum DepositType {
    ERC20,
    ETH
  }

  /**
   * @notice Represents a deposit message.
   * @dev The fields used for `messageHash` generation are:
   * - sender
   * - target
   * - value
   * - nonce
   * - gasLimit
   * - message
   */
  struct DepositMessage {
    address sender; // The address of the sender
    address target; // The target address on the destination chain
    uint256 value; // The value of the deposit
    uint256 nonce; // The nonce for the deposit
    uint256 expiryTime; // The expiry time for the deposit
    bool isCancelled; // Indicates if the deposit is cancelled
    bool isClaimed; // Indicates if the failed deposit is claimed
    address l1DepositRefundAddress; // The address to refund if the deposit is cancelled
    DepositType depositType; // The type of the deposit
    bytes message; // The encoded message data generated by the bridge contract
    INilGasPriceOracle.FeeCreditData feeCreditData;
  }

  /// @notice Gets the current deposit nonce.
  /// @return The current deposit nonce.
  function depositNonce() external view returns (uint256);

  /// @notice Gets the next deposit nonce.
  /// @return The next deposit nonce.
  function getNextDepositNonce() external view returns (uint256);

  /// @notice Gets the deposit type for a given message hash.
  /// @param msgHash The hash of the deposit message.
  /// @return depositType The type of the deposit.
  function getDepositType(bytes32 msgHash) external view returns (DepositType depositType);

  /// @notice Gets the deposit message for a given message hash.
  /// @param msgHash The hash of the deposit message.
  /// @return depositMessage The deposit message details.
  function getDepositMessage(bytes32 msgHash) external view returns (DepositMessage memory depositMessage);

  /// @notice Get the list of authorized bridges
  /// @return The list of authorized bridge addresses.
  function getAuthorizedBridges() external view returns (address[] memory);

  /*//////////////////////////////////////////////////////////////////////////
                           PUBLIC MUTATING FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Send cross chain message from L1 to L2 or L2 to L1.
  /// @param depositType The depositType enum value
  /// @param messageTarget The address of contract/account who receive the message.
  /// @param value The amount of ether passed when call target contract.
  /// @param message The content of the message.
  /// @param l1DepositRefundAddress The address of recipient for the deposit-cancellation or claim failed deposit
  /// @param l2FeeRefundAddress The address of refundRecipient for Fee on L2
  function sendMessage(
    DepositType depositType,
    address messageTarget,
    uint256 value,
    bytes calldata message,
    address l1DepositRefundAddress,
    address l2FeeRefundAddress,
    INilGasPriceOracle.FeeCreditData calldata feeCreditData
  ) external payable;

  /// @notice Send cross chain message from L1 to L2 or L2 to L1.
  /// @param depositType The depositType enum value
  /// @param messageTarget The address of account who receive the message.
  /// @param value The amount of ether passed when call target contract.
  /// @param message The content of the message.
  /// @param refundAddress The address of account who will receive the refunded fee, funds are credited during
  /// finalize-withdrawal, cancel-deposit, claim-failed-deposit
  /// @param feeCreditData feeCreditData struct which holds the l2FeeRefundAddress, nilGasLimit, gasFee data from nilGasPriceOracle
  function sendMessage(
    DepositType depositType,
    address messageTarget,
    uint256 value,
    bytes calldata message,
    address refundAddress,
    INilGasPriceOracle.FeeCreditData calldata feeCreditData
  ) external payable;

  /// @notice Cancels a deposit message.
  /// @param messageHash The hash of the deposit message to cancel.
  function cancelDeposit(bytes32 messageHash) external;

  function claimFailedDeposit(bytes32 messageHash, bytes32[] calldata claimProof) external;

  /*//////////////////////////////////////////////////////////////////////////
                           RESTRICTED FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

  /// @notice Authorize a bridge addresses
  /// @param bridges The array of addresses of the bridges to authorize.
  function authorizeBridges(address[] memory bridges) external;

  /// @notice Authorize a bridge address
  /// @param bridge The address of the bridge to authorize.
  function authorizeBridge(address bridge) external;

  /// @notice Revoke authorization of a bridge address
  /// @param bridge The address of the bridge to revoke.
  function revokeBridgeAuthorization(address bridge) external;

  /// @notice remove a list of messageHash values from the depositMessageQueue.
  /// @dev messages are always popped from the queue in FIFIO Order
  /// @param messageCount number of messages to be removed from the queue
  /// @return messageHashes array of messageHashes start from the head of queue
  function popMessages(uint256 messageCount) external returns (bytes32[] memory);
}
