const { buildSync } = require('esbuild');

module.exports = function () {
  const result = buildSync({ entryPoints: [this.resourcePath], bundle: true, write: false,
    platform: 'browser', format: 'iife', minify: true, metafile: true });
  for (const file of Object.keys(result.metafile.inputs)) this.addDependency(require('path').resolve(file));
  return `export default function createWorker() {
    const url = URL.createObjectURL(new Blob([${JSON.stringify(result.outputFiles[0].text)}], {type:'text/javascript'}));
    try { return new Worker(url); } finally { URL.revokeObjectURL(url); }
  }`;
};
