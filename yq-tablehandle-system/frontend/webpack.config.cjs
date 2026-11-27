const { createWebpackConfig } = require('@pangu/plugin-webpack-preset');

module.exports = async mode => {
  const config = await createWebpackConfig({ mode, context: __dirname, env: process.env.ENV_YQSL || 'dev' });
  config.module.rules.unshift({ resourceQuery: /worker&inline/, enforce: 'pre', use: require.resolve('./build/inline-worker.cjs') });
  return config;
};
