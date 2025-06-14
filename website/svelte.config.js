import adapter from "@sveltejs/adapter-cloudflare";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs/kit/integrations
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  kit: {
    adapter: adapter({
      // Routes will be deployed as Functions by default
      // Static assets will be served from the build directory
      routes: {
        include: ["/*"],
        exclude: ["<build>/*"],
      },
    }),
  },
};

export default config;
