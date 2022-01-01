using System;
using System.Threading.Tasks;
using Microsoft.JSInterop;

namespace PnyxWebAssembly.Client.Services
{
    public class BrowserResizeService
    {
        /// <summary>
        /// Gets or sets the js runtime.
        /// </summary>
        /// <value>
        /// The js runtime.
        /// </value>
        public static IJSRuntime JsRuntime { get; set; }

        /// <summary>
        /// Occurs when [on resize].
        /// </summary>
        public static event Func<Task> OnResize;

        /// <summary>
        /// Called when [browser resize].
        /// </summary>
        [JSInvokable]
        public static async Task OnBrowserResize()
        {
            if (OnResize != null)
            {
                await OnResize.Invoke();
            }
        }

        /// <summary>
        /// Gets the height of the inner.
        /// </summary>
        /// <returns></returns>
        public static async Task<int> GetInnerHeight()
        {
            return await JsRuntime.InvokeAsync<int>("browserResize.getInnerHeight");
        }

        /// <summary>
        /// Gets the width of the inner.
        /// </summary>
        /// <returns></returns>
        public static async Task<int> GetInnerWidth()
        {
            return await JsRuntime.InvokeAsync<int>("browserResize.getInnerWidth");
        }
    }
}
