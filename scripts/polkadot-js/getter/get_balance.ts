import { ApiPromise, WsProvider } from "@polkadot/api";
import { getProvider, getWallets } from "../utils/indexer";

// Put the address of the account you want to fetch info for here

async function fetchAccountInfo() {
  // Initialise the provider to connect to the local node

  // Create the API instance
  const api = await getProvider();

  const wallets = getWallets();
  try {
    // Fetch the account info
    const accountInfo = await api.query.system.account(wallets.alice.address);

    console.log(
      `Account ${wallets.alice.address} info:`,
      accountInfo.toHuman()
    );
  } catch (error) {
    console.error("Error fetching account info:", error);
  }

  try {
    const bobAccountInfo = await api.query.system.account(wallets.bob.address);
    console.log(
      `Account ${wallets.bob.address} info:`,
      bobAccountInfo.toHuman()
    );
  } catch (error) {
    console.error("Error fetching account info:", error);
  } finally {
    // Disconnect the provider when done
    api.disconnect();
  }
}

fetchAccountInfo();
