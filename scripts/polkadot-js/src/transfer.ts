import { ApiPromise } from "@polkadot/api";
import { Keyring } from "@polkadot/keyring";
import { getProvider, getWallets } from "../utils/indexer";

// The amount to transfer
const amount = 1000; // Adjust the amount as needed

async function transferFunds() {
  const api = await getProvider();

  console.log("hello");
  const wallets = getWallets();

  console.log("Alice address: ", wallets.alice.address);
  console.log("Bob address: ", wallets.bob.address);

  // Fetch the next nonce for the Alice's account
  const { nonce } = (await api.query.system.account(
    wallets.alice.address
  )) as any;

  // Construct the transaction
  const transfer = api.tx.balances.transfer(wallets.bob.address, amount);

  console.log(
    `Transferring ${amount} from ${wallets.alice.address} to ${wallets.bob.address}`
  );
  console.log(`Nonce: ${nonce}`);

  // Sign and send the transaction, and subscribe to its status updates
  const unsub = await transfer.signAndSend(
    wallets.alice,
    ({ status, events }) => {
      if (status.isInBlock) {
        console.log(`Transaction included at blockHash ${status.asInBlock}`);
      } else if (status.isFinalized) {
        console.log(`Transaction finalized at blockHash ${status.asFinalized}`);
        events.forEach(({ event: { data, method, section }, phase }) => {
          console.log(`\t' ${phase}: ${section}.${method} ${data}`);
        });

        // Once finalized, we can unsubscribe from further updates
        unsub();
        // Disconnect from the provider
        api.disconnect();
      }
    }
  );
}

transferFunds();
