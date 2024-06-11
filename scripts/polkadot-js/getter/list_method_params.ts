import { ApiPromise } from "@polkadot/api";
import { getProvider } from "../utils/indexer";

async function run() {
  const api = await getProvider();
  await listTxMethods(api);
}
type MetadataV14 = {
  magicNumber: string;
  metadata: {
    v14: {
      lookup: {
        types: [
          {
            id: string;
            type: {
              path: string[];
              params: object[];
              def: object;
              docs: string[];
            };
          }
        ];
      };
      pallets: Array<{
        name: string;
        calls?: Array<{
          name: string;
          args: Array<{
            name: string;
            type: string | number; // Depending on how types are represented, you might need to adjust this
          }>;
        }>;
      }>;
      extrinsic: object;
      type: string;
    };
  };
};

async function listTxMethods(api: ApiPromise) {
  console.log("\nTransaction Methods:");
  const metadata = await api.rpc.state.getMetadata();

  const metadataV14 = metadata.toJSON() as {
    magicNumber: string;
    metadata: {
      v14: {
        lookup: {
          types: [
            {
              id: string;
              type: {
                path: string[];
                params: object[];
                def: object;
                docs: string[];
              };
            }
          ];
        };
        pallets: Array<any>;
        extrinsic: object;
        type: string;
      };
    };
  };

  console.log("pallets: ", metadataV14.metadata.v14.pallets);
  // Usage example, assuming you have metadataV14 of type MetadataV14
  const ibcModule = metadataV14.metadata.v14.pallets.find(
    (pallet) => pallet.name === "Ibc"
  );

  if (ibcModule && ibcModule.calls) {
    const transferMethod = ibcModule.calls.find(
      (call: any) => call.name === "transfer"
    );
    if (transferMethod) {
      console.log(`Parameters for ibc.transfer:`);
      transferMethod.args.forEach((arg: any) => {
        console.log(`${arg.name}: ${arg.type}`);
      });
    }
  }
}

run().catch(console.error);
