import { BaseClient } from "../clients/BaseClient.js";
import type { IClientBaseConfig } from "../clients/types/Configs.js";
import type { Hex } from "../types/Hex.js";
import type { ContractData, Location, TransactionData } from "./CometaTypes.js";

/**
 * The type representing the config for the Cometa service client.
 */
type CometaServiceConfig = IClientBaseConfig;

/**
 * CometaService is a client that interacts with the Cometa service.
 * Cometa service is used to store contract metadata: source code, ABI, etc.
 * @class CometaService
 * @extends BaseClient
 * @example
 * import { CometaService } from '@nilfoundation/niljs';
 *
 * const service = new CometaService({
 *   transport: new HttpTransport({
 *     endpoint: COMETA_ENDPOINT,
 *   }),
 * });
 */
class CometaService extends BaseClient {
  // biome-ignore lint/complexity/noUselessConstructor: may be useful in the future
  constructor(config: CometaServiceConfig) {
    super(config);
  }

  /**
   * Returns the contract metadata.
   * @param address - Address of the contract.
   * @returns The contract metadata.
   */
  public async getContract(address: Hex) {
    return await this.request<ContractData>({
      method: "cometa_getContract",
      params: [address],
    });
  }

  /**
   * Returns the contract metadata.
   * @param address - Address of the contract.
   * @param pc - Program counter.
   * @returns The contract metadata.
   */
  public async getLocation(address: Hex, pc: number) {
    return await this.request<Location>({
      method: "cometa_getLocation",
      params: [address, pc],
    });
  }

  /**
   * Compiles the contract.
   * @param inputJson - The JSON input.
   * @returns The contract metadata.
   */
  public async compileContract(inputJson: string | object) {
    const inputJsonString = typeof inputJson === "string" ? inputJson : JSON.stringify(inputJson);

    return await this.request<ContractData>({
      method: "cometa_compileContract",
      params: [inputJsonString],
    });
  }

  /**
   * Register the contract by compilation result.
   * @param contractData - The contract data.
   * @param address - Address of the contract.
   */
  public async registerContractData(contractData: ContractData, address: Hex) {
    return await this.request({
      method: "cometa_registerContractData",
      params: [contractData, address],
    });
  }

  /**
   * Register the contract.
   * @param inputJson - The JSON input for compiler.
   * @param address - Address of the contract.
   */
  public async registerContract(inputJson: string | object, address: Hex) {
    const inputJsonString = typeof inputJson === "string" ? inputJson : JSON.stringify(inputJson);

    return await this.request({
      method: "cometa_registerContract",
      params: [inputJsonString, address],
    });
  }

  /**
   * Returns the abi of the contract.
   * @param address - Address of the contract.
   * @returns Abi of the contract.
   */
  public async getAbi(address: Hex) {
    return await this.request<string>({
      method: "cometa_getAbi",
      params: [address],
    });
  }

  /**
   * Returns the source code of the contract by address.
   * @param address - Address of the contract.
   * @returns The source code of the contract.
   */
  public async getSourceCode(address: Hex) {
    return await this.request<Record<string, string>>({
      method: "cometa_getSourceCode",
      params: [address],
    });
  }

  /**
   * Accepts an array of transaction data and returns an array of decoded function names called by the transactions.
   * @param data - The data to decode.
   * @returns The decoded data.
   */
  public async decodeTransactionsCallData(data: TransactionData[]) {
    return await this.request<string[]>({
      method: "cometa_decodeTransactionsCallData",
      params: [data],
    });
  }
}

export { CometaService };
