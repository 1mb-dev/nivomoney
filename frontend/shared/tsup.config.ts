import { defineConfig } from 'tsup';

export default defineConfig([
  // Main utilities bundle (non-React)
  {
    entry: { index: 'src/index.ts' },
    format: ['cjs', 'esm'],
    dts: true,
    clean: true,
    outDir: 'dist',
  },
  // Components bundle (React components)
  {
    entry: { components: 'components/index.ts' },
    format: ['cjs', 'esm'],
    dts: true,
    outDir: 'dist',
    external: ['react', 'react-dom', 'react-router-dom'],
    esbuildOptions(options) {
      options.jsx = 'automatic';
    },
  },
  // Utils bundle (cn function)
  {
    entry: { utils: 'lib/utils.ts' },
    format: ['cjs', 'esm'],
    dts: true,
    outDir: 'dist',
  },
]);
