import { ethers, network } from 'hardhat';
import { Contract } from 'ethers';
import * as fs from 'fs';
import * as path from 'path';
import { loadConfig, isValidAddress } from '../../deploy/config/config-helper';

const abiPath = path.join(
    __dirname,
    '../../artifacts/contracts/NilAccessControl.sol/NilAccessControl.json',
);
const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8')).abi;

// npx hardhat run scripts/access-control/has-a-role.ts --network sepolia
export async function hasRole(roleHash: string, account: string) {
    const networkName = network.name;
    const config = loadConfig(networkName);

    if (!isValidAddress(config.nilRollupProxy)) {
        throw new Error('Invalid nilRollupProxy address in config');
    }

    const [signer] = await ethers.getSigners();

    const nilRollupInstance = new ethers.Contract(
        config.nilRollupProxy,
        abi,
        signer,
    ) as Contract;

    return await nilRollupInstance.hasRole(roleHash, account);
}
