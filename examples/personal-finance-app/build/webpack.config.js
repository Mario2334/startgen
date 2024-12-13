const path = require('path');

module.exports = {
  entry: './renderer/renderer.js',
  output: {
    filename: 'bundle.js',
    path: path.resolve(__dirname, 'dist')
  },
  mode: 'production'
};