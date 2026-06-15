import type { CodegenConfig } from "@graphql-codegen/cli";

const config: CodegenConfig = {
  schema: "../backend/graph/schema/*.graphqls",
  documents: ["src/**/*.graphql"],
  ignoreNoDocuments: true,
  generates: {
    "src/graphql/generated/graphql.ts": {
      plugins: [
        "typescript",
        "typescript-operations",
        "typescript-react-apollo",
      ],
      config: {
        withHooks: true,
        reactApolloVersion: 3,
        enumsAsTypes: true,
        scalars: {
          UUID: "string",
          DateTime: "string",
        },
      },
    },
  },
};

export default config;
