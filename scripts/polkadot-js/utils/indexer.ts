import { ApiPromise, WsProvider } from "@polkadot/api";
import { Keyring } from "@polkadot/keyring";
import { KeyringPair } from "@polkadot/keyring/types";

export type TestWallets = {
  alice: KeyringPair;
  bob: KeyringPair;
};

export const getProvider = async (): Promise<ApiPromise> => {
  // Initialise the provider to connect to the local node
  const wsEndpoint = "ws://127.0.0.1:9988";
  const provider = new WsProvider(wsEndpoint);

  console.log(`connection to provider at ${wsEndpoint}`);

  // Create the API and wait until ready
  const api = await ApiPromise.create({ provider });

  return api;
};

export const getWallets = (): TestWallets => {
  // Add Alice to our keyring with a well-known development mnemonic
  const keyring = new Keyring({ type: "sr25519" });
  const alice = keyring.addFromUri("//Alice");
  const bob = keyring.addFromUri("//Bob");

  return { alice, bob };
};
