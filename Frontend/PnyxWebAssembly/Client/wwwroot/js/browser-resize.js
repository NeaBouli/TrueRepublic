window.browserResize = {
    getInnerHeight: function () {
        return window.innerHeight;
    },
    getInnerWidth: function () {
        return window.innerWidth;
    },
    registerResizeCallback: function () {
        window.addEventListener("resize", browserResize.resized);
    },
    resized: function () {
        // window.DotNet.invokeMethod("BrowserResize", 'OnBrowserResize');
        window.DotNet.invokeMethodAsync("PnyxWebAssembly.Client", 'OnBrowserResize').then(data => data);
    }
}