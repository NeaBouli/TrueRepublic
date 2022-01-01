using System;
using System.Security.Claims;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Components;
using Microsoft.AspNetCore.Components.Authorization;
using Microsoft.JSInterop;
using PnyxWebAssembly.Client.Services;
using PnyxWebAssembly.Client.Shared;

namespace PnyxWebAssembly.Client.Pages
{
    /// <summary>
    /// Implementation of the index code behind class
    /// </summary>
    /// <seealso cref="Microsoft.AspNetCore.Components.ComponentBase" />
    public partial class Index : IDisposable
    {
        /// <summary>
        /// Gets or sets the js runtime.
        /// </summary>
        /// <value>
        /// The js runtime.
        /// </value>
        [Inject] 
        private IJSRuntime JsRuntime { get; set; }

        /// <summary>
        /// Gets or sets the authentication state provider.
        /// </summary>
        /// <value>
        /// The authentication state provider.
        /// </value>
        [Inject]
        private AuthenticationStateProvider AuthenticationStateProvider { get; set; }

        [CascadingParameter]
        public MainLayout Layout { get; set; }

        /// <summary>
        /// Gets or sets the height.
        /// </summary>
        /// <value>
        /// The height.
        /// </value>
        public string Height { get; set; }

        /// <summary>
        /// Gets or sets the width.
        /// </summary>
        /// <value>
        /// The width.
        /// </value>
        public string Width { get; set; }

        /// <summary>
        /// Gets or sets the name of the user.
        /// </summary>
        /// <value>
        /// The name of the user.
        /// </value>
        public string UserName { get; set; }

        /// <summary>
        /// Method invoked when the component is ready to start, having received its
        /// initial parameters from its parent in the render tree.
        /// Override this method if you will perform an asynchronous operation and
        /// want the component to refresh when that operation is completed.
        /// </summary>
        protected override async Task OnInitializedAsync()
        {
            BrowserResizeService.JsRuntime = JsRuntime;

            BrowserResizeService.OnResize += BrowserHasResized;

            await JsRuntime.InvokeAsync<object>("browserResize.registerResizeCallback");

            await GetDimensions();

            AuthenticationState authState = await AuthenticationStateProvider.GetAuthenticationStateAsync();
            ClaimsPrincipal user = authState.User;

            UserName = user?.Identity?.Name;

            if (!string.IsNullOrEmpty(UserName))
            {
                // TODO: call get total balance service here
                Layout.TotalBalance++;
            }
        }

        /// <summary>
        /// Browsers the has resized.
        /// </summary>
        private async Task BrowserHasResized()
        {
            await GetDimensions();

            StateHasChanged();
        }

        /// <summary>
        /// Gets the dimensions.
        /// </summary>
        private async Task GetDimensions()
        {
            int height = await BrowserResizeService.GetInnerHeight();
            int width = await BrowserResizeService.GetInnerWidth();

            Height = $@"{height - 85}px";
            Width = $@"{width}px";
        }

        /// <summary>
        /// Performs application-defined tasks associated with freeing, releasing, or resetting unmanaged resources.
        /// </summary>
        public void Dispose()
        {
            BrowserResizeService.OnResize -= BrowserHasResized;
        }
    }
}
