import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

// Dev
// import adapter from "@sveltejs/adapter-auto";

// Prod
import adapter from "@sveltejs/adapter-static";


// Dev
// const config = {
// 	preprocess: vitePreprocess(),
// 	kit: {
// 		adapter: adapter()
// 	}
// };

// Prod
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			fallback: "index.html"
		})
	}
};

export default config;
