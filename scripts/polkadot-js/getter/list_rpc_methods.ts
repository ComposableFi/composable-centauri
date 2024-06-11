import { ApiPromise, WsProvider } from "@polkadot/api";
import { getProvider, getWallets } from "../utils/indexer";

// Put the address of the account you want to fetch info for here

async function run() {
  // Initialise the provider to connect to the local node

  // Create the API instance
  const api = await getProvider();

  listTxMethods(api);
}

run();

function listTxMethods(api: ApiPromise) {
  console.log("\nTransaction Methods:");
  Object.keys(api.tx).forEach((module) => {
    Object.keys(api.tx[module]).forEach((method) => {
      console.log(`${module}.${method}`);
    });
  });
}

function listQueryMethods(api: ApiPromise) {
  console.log("\nQuery Methods:");
  Object.keys(api.query).forEach((module) => {
    Object.keys(api.query[module]).forEach((method) => {
      console.log(`${module}.${method}`);
    });
  });
}
