const webpack = require("webpack");

module.exports = function override(config) {
    config.resolve.fallback = {
        ...config.resolve.fallback,
        crypto: require.resolve("crypto-browserify"),
        stream: require.resolve("stream-browserify"),
        buffer: require.resolve("buffer/"),
        vm: false,
        path: false,
        os: false,
        fs: false,
        http: false,
        https: false,
        zlib: false,
        url: false,
        assert: false,
    };
    config.plugins = [
        ...config.plugins,
        new webpack.ProvidePlugin({
            Buffer: ["buffer", "Buffer"],
        }),
    ];
    return config;
};
