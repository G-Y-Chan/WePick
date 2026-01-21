// metro.config.js
const { getDefaultConfig } = require("expo/metro-config"); // use expo since you're on a dev client

const config = getDefaultConfig(__dirname);

config.transformer = {
  ...config.transformer,
  unstable_allowRequireContext: true,
};

module.exports = config;
