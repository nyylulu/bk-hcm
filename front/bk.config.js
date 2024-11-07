const webpack = require('webpack');
const CopyWebpackPlugin = require('copy-webpack-plugin');
const fs = require('fs');
const { resolve, join } = require('path');
const replaceStaticUrlPlugin = require('./replace-static-url-plugin');
const isModeProduction = process.env.NODE_ENV === 'production';
const indexPath = isModeProduction ? './index.html' : './index-dev.html';
const env = require('./env')();
const apiMocker = require('./mock-server.js');
module.exports = {
  appConfig() {
    return {
      indexPath,
      mainPath: './src/main.ts',
      publicPath: env.publicPath,
      outputDir: env.outputDir,
      assetsDir: env.assetsDir,
      minChunkSize: 10000,
      // pages: {
      //   main: {
      //     entry: './src/main.ts',
      //     filename: 'index.html'
      //   },
      // },
      // needSplitChunks: false,
      css: {
        loaderOptions: {
          scss: {
            additionalData: '@import "./src/style/variables.scss";',
          },
        },
      },
      devServer: {
        host: env.DEV_HOST,
        port: 5000,
        historyApiFallback: true,
        disableHostCheck: true,
        before(app) {
          apiMocker(app, {
            // watch: [
            //   '/mock/api/v4/organization/user_info/',
            //   '/mock/api/v4/add/',
            //   '/mock/api/v4/get/',
            //   '/mock/api/v4/sync/',
            //   '/mock/api/v4/cloud/public_images/list/'
            // ],
            api: resolve(__dirname, './mock/api.ts'),
          });
        },
        proxy: {},
      },
    };
  },
  configureWebpack(_webpackConfig) {
    webpackConfig = _webpackConfig;
    const extensions = ['.js', '.vue', '.json', '.ts', '.tsx'];
    webpackConfig.plugins.push(
      new replaceStaticUrlPlugin(),
      new webpack.NormalModuleReplacementPlugin(/\.plugin(\.\w+)?$/, function (resource) {
        // 获取文件的绝对路径
        const absPath = resolve(resource.context, resource.request);

        // 内部插件后缀
        const internalPluginSuffix = '-internal.plugin';

        // 构造内部插件和默认插件的基本路径（不带扩展名）
        const internalPluginBase = absPath.replace(/\.plugin/, internalPluginSuffix);

        // 定义一个辅助函数来检查文件是否存在
        const fileExists = (basePath) => {
          return extensions.some((ext) => fs.existsSync(`${basePath}${ext}`));
        };

        // 检查 -internal.plugin 文件是否存在
        if (env.isInternal && fileExists(internalPluginBase)) {
          // 如果内部插件存在，使用它
          resource.request = internalPluginBase;
        } else {
          // 如果没有任何插件文件存在，发出警告并保留原始 request
          console.error(`Warning: No suitable file found for ${internalPluginBase}. Using default plugin file.`);
          resource.request = absPath;
        }

        // 如果 createData 存在，也同步更新它的 request
        if (resource.createData) {
          resource.createData.request = resource.request;
        }
      }),
    );
    webpackConfig.plugins.push(
      new CopyWebpackPlugin({
        patterns: [
          {
            from: resolve('static/image'),
            to: resolve('dist'),
            globOptions: {
              ignore: [
                // 忽略所有 HTML 文件，如果有的话
                '**/*.html',
              ],
            },
          },
          {
            from: 'static/*.html', // 只匹配 static 目录下的 HTML 文件
            to: '[name][ext]', // 保持原文件名
          },
        ],
      }),
    );

    // webpackConfig.externals = {
    //   'axios':'axios',
    //   'dayjs':'dayjs',
    // }
    webpackConfig.resolve = {
      ...webpackConfig.resolve,
      symlinks: false,
      extensions,
      alias: {
        ...webpackConfig.resolve?.alias,
        // extensions: ['.js', '.jsx', '.ts', '.tsx'],
        '@': resolve(__dirname, './src'),
        '@static': resolve(__dirname, './static'),
        '@charts': resolve(__dirname, './src/plugins/charts'),
        '@datasource': resolve(__dirname, './src/plugins/datasource'),
        '@modules': resolve(__dirname, './src/store/modules'),
        '@pluginHandler': resolve(__dirname, `./src/plugin-handler${env.isInternal ? '/bcc' : ''}`),
      },
    };
  },
};
