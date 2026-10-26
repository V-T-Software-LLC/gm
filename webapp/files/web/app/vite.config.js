import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

// in development the go server loads scripts from this dev server.
// builds go to ../static with stable file names so templ can link them directly.
export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  server: {
    port: 5173,
    strictPort: true,
    origin: 'http://localhost:5173',
  },
  build: {
    outDir: '../static',
    emptyOutDir: true,
    rolldownOptions: {
      input: 'src/main.js',
      output: {
        entryFileNames: '[name].js',
        assetFileNames: '[name][extname]',
      },
    },
  },
})
