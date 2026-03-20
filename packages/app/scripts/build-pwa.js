import { generateSW } from "workbox-build";

await generateSW({
  globDirectory: "dist",
  globPatterns: ["**/*.{js,css,html,png,svg,json}"],
  swDest: "dist/service-worker.js",
  clientsClaim: true,
  skipWaiting: true
});
