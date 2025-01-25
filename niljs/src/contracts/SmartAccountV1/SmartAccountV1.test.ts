import { PublicClient } from "../../clients/index.js";
import { type SmartAccountV1Config, generateRandomPrivateKey } from "../../index.js";
import { LocalECDSAKeySigner } from "../../signers/LocalECDSAKeySigner.js";
import { MockTransport } from "../../transport/MockTransport.js";
import { SmartAccountV1 } from "./SmartAccountV1.js";

const signer = new LocalECDSAKeySigner({
  privateKey: generateRandomPrivateKey(),
});
const pubkey = signer.getPublicKey();

const fn = vi.fn();
fn.mockReturnValue({});

const client = new PublicClient({
  transport: new MockTransport(fn),
  shardId: 1,
});

test("Smart account creation test with salt and no salt", async ({ expect }) => {
  describe("empty smart account creation", () => {
    expect(() => new SmartAccountV1({} as SmartAccountV1Config)).toThrowError();
  });
  describe("smart account creation with address and salt", () => {
    expect(
      () =>
        // @ts-ignore - Testing invalid input
        new SmartAccountV1({
          pubkey: pubkey,
          salt: 100n,
          shardId: 1,
          client,
          signer,
          address: SmartAccountV1.calculateSmartAccountAddress({
            pubKey: pubkey,
            shardId: 1,
            salt: 100n,
          }),
        }),
    ).toThrowError();
  });

  expect(
    () =>
      new SmartAccountV1({
        pubkey: pubkey,
        salt: 100n,
        shardId: 1,
        client,
        signer,
      }),
  ).toBeDefined();

  expect(
    () =>
      new SmartAccountV1({
        pubkey: pubkey,
        client,
        signer,
        address: SmartAccountV1.calculateSmartAccountAddress({
          pubKey: pubkey,
          shardId: 1,
          salt: 100n,
        }),
      }),
  ).toBeDefined();

  expect(
    () =>
      new SmartAccountV1({
        pubkey: pubkey,
        client,
        signer,
        address: SmartAccountV1.calculateSmartAccountAddress({
          pubKey: pubkey,
          shardId: 1,
          salt: 100n,
        }),
      }),
  ).toBeDefined();
});

test("Smart account self deploy test", async ({ expect }) => {
  const smartAccount = new SmartAccountV1({
    pubkey: pubkey,
    client,
    signer,
    address: SmartAccountV1.calculateSmartAccountAddress({
      pubKey: pubkey,
      shardId: 1,
      salt: 100n,
    }),
  });

  await expect(async () => {
    await smartAccount.selfDeploy(true);
  }).rejects.toThrowError();
});

test("Deploy through smart account", async ({ expect }) => {
  const fn = vi.fn();

  const client = new PublicClient({
    transport: new MockTransport(fn),
  });
  const smartAccount = new SmartAccountV1({
    pubkey: pubkey,
    client,
    signer,
    address: SmartAccountV1.calculateSmartAccountAddress({
      pubKey: pubkey,
      shardId: 1,
      salt: 100n,
    }),
  });
  await smartAccount.deployContract({
    abi: [],
    bytecode: "0x222222222222222222222222222222222222222222222222222222222222222222",
    args: [],
    chainId: 1,
    seqno: 1,
    salt: 100n,
    shardId: 1,
    value: 100n,
    feeCredit: 100_000n,
  });
  expect(fn.mock.calls).toHaveLength(1);
  expect(fn.mock.calls[0][0].method).toBe("eth_sendRawTransaction");
  expect(fn.mock.calls[0][0].params[0]).toContain([
    "222222222222222222222222222222222222222222222222222222222222222222",
  ]);
});
