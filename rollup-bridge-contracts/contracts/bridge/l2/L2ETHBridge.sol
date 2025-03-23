// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { PausableUpgradeable } from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { NilAccessControlUpgradeable } from "../../NilAccessControlUpgradeable.sol";
import { ReentrancyGuardUpgradeable } from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import { AccessControlEnumerableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/extensions/AccessControlEnumerableUpgradeable.sol";
import { NilConstants } from "../../common/libraries/NilConstants.sol";
import { IERC165 } from "@openzeppelin/contracts/utils/introspection/IERC165.sol";
import { EnumerableSet } from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import { IL2Bridge } from "./interfaces/IL2Bridge.sol";
import { AddressChecker } from "../../common/libraries/AddressChecker.sol";
import { IL1ETHBridge } from "../l1/interfaces/IL1ETHBridge.sol";
import { IL2ETHBridge } from "./interfaces/IL2ETHBridge.sol";
import { IL2ETHBridgeVault } from "./interfaces/IL2ETHBridgeVault.sol";
import { IL2BridgeMessenger } from "./interfaces/IL2BridgeMessenger.sol";
import { IL2BridgeRouter } from "./interfaces/IL2BridgeRouter.sol";

contract L2ETHBridge is
  OwnableUpgradeable,
  PausableUpgradeable,
  NilAccessControlUpgradeable,
  ReentrancyGuardUpgradeable,
  IL2ETHBridge
{
  using EnumerableSet for EnumerableSet.AddressSet;
  using AddressChecker for address;

  /*//////////////////////////////////////////////////////////////////////////
                             STATE-VARIABLES   
    //////////////////////////////////////////////////////////////////////////*/

  address public override counterpartyBridge;

  address public override messenger;

  address public override router;

  IL2ETHBridgeVault public override l2ETHBridgeVault;

  /// @notice Mapping from Enshrined-Token-Address from nil-shard to ERC20-Token-Address from L1
  mapping(address => address) public tokenMapping;

  /// @dev The storage slots for future usage.
  uint256[50] private __gap;

  /*//////////////////////////////////////////////////////////////////////////
                                    CONSTRUCTOR
    //////////////////////////////////////////////////////////////////////////*/

  /// @custom:oz-upgrades-unsafe-allow constructor
  constructor() {
    _disableInitializers();
  }

  /*//////////////////////////////////////////////////////////////////////////
                                    INITIALIZER
    //////////////////////////////////////////////////////////////////////////*/

  function initialize(
    address ownerAddress,
    address adminAddress,
    address routerAddress,
    address messengerAddress,
    address l2ETHBridgeVaultAddress
  ) public initializer {
    // Validate input parameters
    if (ownerAddress == address(0)) {
      revert ErrorInvalidOwner();
    }

    if (adminAddress == address(0)) {
      revert ErrorInvalidDefaultAdmin();
    }

    // Initialize the Ownable contract with the owner address
    OwnableUpgradeable.__Ownable_init(ownerAddress);

    // Initialize the Pausable contract
    PausableUpgradeable.__Pausable_init();

    // Initialize the AccessControlEnumerable contract
    __AccessControlEnumerable_init();

    ReentrancyGuardUpgradeable.__ReentrancyGuard_init();

    _setRouter(routerAddress);
    _setMessenger(messengerAddress);
    _setL2ETHBridgeVault(l2ETHBridgeVaultAddress);

    // Set role admins
    // The OWNER_ROLE is set as its own admin to ensure that only the current owner can manage this role.
    _setRoleAdmin(NilConstants.OWNER_ROLE, NilConstants.OWNER_ROLE);

    // The DEFAULT_ADMIN_ROLE is set as its own admin to ensure that only the current default admin can manage this
    // role.
    _setRoleAdmin(DEFAULT_ADMIN_ROLE, NilConstants.OWNER_ROLE);

    // Grant roles to defaultAdmin and owner
    // The DEFAULT_ADMIN_ROLE is granted to both the default admin and the owner to ensure that both have the
    // highest level of control.
    // The OWNER_ROLE is granted to the owner to ensure they have the highest level of control over the contract.
    _grantRole(NilConstants.OWNER_ROLE, ownerAddress);
    _grantRole(DEFAULT_ADMIN_ROLE, adminAddress);
  }

  /*//////////////////////////////////////////////////////////////////////////
                                    MODIFIERS
    //////////////////////////////////////////////////////////////////////////*/

  modifier onlyMessenger() {
    // check caller is l2-bridge-messenger
    if (msg.sender != address(messenger)) {
      revert ErrorCallerIsNotMessenger();
    }
    _;
  }

  /*//////////////////////////////////////////////////////////////////////////
                                    PUBLIC MUTATION FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

  function finaliseETHDeposit(
    address depositorAddress,
    address payable depositRecipient,
    address feeRefundRecipient,
    uint256 amount
  ) public payable onlyMessenger {
    // get recipient balance before ETH transfer
    uint256 befBalance = feeRefundRecipient.balance;

    // call sendEth on L2ETHBridgeVault
    l2ETHBridgeVault.transferETH(depositRecipient, amount);

    // check for balance change of recipient
    if (feeRefundRecipient.balance - befBalance != amount) {
      revert ErrorIncompleteETHDeposit();
    }

    // emit FinalisedETHDepositEvent
    emit FinaliseETHDeposit(depositorAddress, depositRecipient, amount);
  }

  /*//////////////////////////////////////////////////////////////////////////
                         RESTRICTED FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

  /// @inheritdoc IL2Bridge
  function setRouter(address routerAddress) external override onlyAdmin {
    _setRouter(routerAddress);
  }

  function _setRouter(address routerAddress) internal {
    if (!routerAddress.isContract() || !IERC165(routerAddress).supportsInterface(type(IL2BridgeRouter).interfaceId)) {
      revert ErrorInvalidRouter();
    }
    emit L2BridgeRouterSet(router, routerAddress);
    router = routerAddress;
  }

  /// @inheritdoc IL2Bridge
  function setMessenger(address messengerAddress) external override onlyAdmin {
    _setMessenger(messengerAddress);
  }

  function _setMessenger(address messengerAddress) internal {
    if (
      !messengerAddress.isContract() ||
      !IERC165(messengerAddress).supportsInterface(type(IL2BridgeMessenger).interfaceId)
    ) {
      revert ErrorInvalidMessenger();
    }
    emit L2BridgeMessengerSet(messenger, messengerAddress);
    messenger = messengerAddress;
  }

  /// @inheritdoc IL2Bridge
  function setCounterpartyBridge(address counterpartyBridgeAddress) external override onlyAdmin {
    _setCounterpartyBridge(counterpartyBridgeAddress);
  }

  function _setCounterpartyBridge(address counterpartyBridgeAddress) internal {
    if (
      counterpartyBridgeAddress != address(0) &&
      (!counterpartyBridgeAddress.isContract() ||
        !IERC165(counterpartyBridgeAddress).supportsInterface(type(IL1ETHBridge).interfaceId))
    ) {
      revert ErrorInvalidCounterpartyBridge();
    }
    emit CounterpartyBridgeSet(counterpartyBridge, counterpartyBridgeAddress);
    counterpartyBridge = counterpartyBridgeAddress;
  }

  /// @inheritdoc IL2ETHBridge
  function setL2ETHBridgeVault(address l2ETHBridgeVaultAddress) external override onlyAdmin {
    _setL2ETHBridgeVault(l2ETHBridgeVaultAddress);
  }

  function _setL2ETHBridgeVault(address l2ETHBridgeVaultAddress) internal {
    if (
      !l2ETHBridgeVaultAddress.isContract() ||
      !IERC165(l2ETHBridgeVaultAddress).supportsInterface(type(IL2ETHBridgeVault).interfaceId)
    ) {
      revert ErrorInvalidEthBridgeVault();
    }

    emit L2ETHBridgeVaultSet(address(l2ETHBridgeVault), l2ETHBridgeVaultAddress);
    l2ETHBridgeVault = IL2ETHBridgeVault(l2ETHBridgeVaultAddress);
  }

  /// @inheritdoc IL2Bridge
  function setPause(bool _status) external override onlyOwner {
    if (_status) {
      _pause();
    } else {
      _unpause();
    }
  }

  /// @inheritdoc IL2Bridge
  function transferOwnershipRole(address newOwner) external override onlyOwner {
    _revokeRole(NilConstants.OWNER_ROLE, owner());
    super.transferOwnership(newOwner);
    _grantRole(NilConstants.OWNER_ROLE, newOwner);
  }

  /// @inheritdoc IERC165
  function supportsInterface(bytes4 interfaceId) public view override returns (bool) {
    return
      interfaceId == type(IL2ETHBridge).interfaceId ||
      interfaceId == type(IL2Bridge).interfaceId ||
      super.supportsInterface(interfaceId);
  }
}
