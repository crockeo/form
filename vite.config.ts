import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";

export default defineConfig({
	plugins: [solidPlugin(), tailwindcss()],
	server: {
		port: 3000,
		proxy: {
			"/api": "http://127.0.0.1:8000",
		},
	},
	build: {
		target: "esnext",
	},
});
