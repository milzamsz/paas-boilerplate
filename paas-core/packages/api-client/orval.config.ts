import { defineConfig } from "orval";

export default defineConfig({
  paas: {
    input: {
      target: "../../apps/api/docs/openapi.json",
    },
    output: {
      target: "./src/api.ts",
      client: "axios",
      mode: "single",
      override: {
        mutator: {
          path: "./src/custom-instance.ts",
          name: "customInstance",
        },
      },
    },
  },
});
