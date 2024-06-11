import { ApiPromise, WsProvider } from "@polkadot/api";
import { Keyring } from "@polkadot/keyring";
import BN from "bn.js";
import { KeyringPair } from "@polkadot/keyring/types";
import { getProvider, getWallets } from "../utils/indexer";

async function sendIbcFundsTx(
  api: ApiPromise,
  senderKeypair: KeyringPair,
  channelID: string,
  amount: { denom: string; amount: string; address: string },
  options: any
) {
  {
    // Ensure the API is connected
    if (!api.isConnected) {
      await api.connect();
    }

    // Calculate the timestamp for 5 minutes into the future
    const fiveMinutes = 5 * 60 * 1000; // 5 minutes in milliseconds
    const futureTimestamp = new Date().getTime() + fiveMinutes; // Current time + 5 minutes

    const substrateFutureTimestamp = api.createType("u64", futureTimestamp);

    // dont have to convert
    const to = { Raw: amount.address };

    const assetNum = 1;
    const sourceChannel = 0;
    const timeout = {
      Offset: {
        timestamp: api.createType("Option<u64>", substrateFutureTimestamp), // or provide a specific timestamp offset
      },
    };

    // Construct paramters
    const params = {
      to,
      source_channel: sourceChannel,
      timeout,
    };

    const assetId = new BN(assetNum);
    const amountBN = new BN(amount.amount, 10);
    const memo = null;

    // Make the call to ibc.transfer with the transferObj
    const call = api.tx.ibc.transfer(params, assetId, amountBN, memo);
    // Sign and send the transaction
    return await new Promise((resolve, reject) => {
      call
        .signAndSend(
          senderKeypair,
          { nonce: -1 },
          ({ status, dispatchError }) => {
            if (status.isInBlock || status.isFinalized) {
              if (dispatchError) {
                if (dispatchError.isModule) {
                  // For module errors, we have the section indexed, lookup
                  const decoded = api.registry.findMetaError(
                    dispatchError.asModule
                  );
                  const { docs, name, section } = decoded;
                  reject(new Error(`${section}.${name}: ${docs.join(" ")}`));
                } else {
                  // Other, CannotLookup, BadOrigin, no extra info
                  reject(new Error(dispatchError.toString()));
                }
              } else {
                resolve(status.asFinalized.toString());
              }
            }
          }
        )
        .catch(reject);
    });
  }
}
// Example usage
async function main() {
  const api = await getProvider();
  const wallets = getWallets();
  const senderKeypair = wallets.alice;

  const channelID = "0";
  const amount = {
    denom: "1",
    amount: "1000000000000000",
    address: "pica1hj5fveer5cjtn4wd6wstzugjfdxzl0xpas3hgy",
  };

  const options = {};

  try {
    const hash = await sendIbcFundsTx(
      api,
      senderKeypair,
      channelID,
      amount,
      options
    );
    console.log("Transaction hash:", hash);
  } catch (error) {
    console.error("Error sending IBC funds:", error);
  }
}

main().catch(console.error);
