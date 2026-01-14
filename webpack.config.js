const path = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

module.exports = {
  mode: 'development', // Set to 'production' for production builds
  entry: './src/static/js/index.js',
  output: {
    path: path.resolve(__dirname, 'src/static/dist'),
    filename: 'main.js',
    publicPath: '/static/dist/', // Public path for the bundled assets
  },
  plugins: [new MiniCssExtractPlugin({
    filename: 'main.css',
  })],
  module: {
    rules: [
      {
        test: /\.css$/i,
        use: [MiniCssExtractPlugin.loader, 'css-loader'],
      },
      // Rule for handling JavaScript files
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: 'babel-loader', // You might need babel-loader for older browser compatibility
          options: {
            presets: ['@babel/preset-env'],
          },
        },
      },
    ],
  },
  resolve: {
    modules: [path.resolve(__dirname, 'node_modules')],
  },
};
