using Microsoft.AspNetCore.Components.WebAssembly.Authentication;
using Microsoft.AspNetCore.Components.WebAssembly.Hosting;
using Microsoft.Extensions.DependencyInjection;
using System;
using System.Threading.Tasks;
using MudBlazor.Services;
using PnyxWebAssembly.Client.Services;

namespace PnyxWebAssembly.Client
{
    /// <summary>
    /// Implementation of the program
    /// </summary>
    public class Program
    {
        /// <summary>
        /// Defines the entry point of the application.
        /// </summary>
        /// <param name="args">The arguments.</param>
        public static async Task Main(string[] args)
        {
            var builder = WebAssemblyHostBuilder.CreateDefault(args);
            builder.RootComponents.Add<App>("#app");

            builder.Services.AddHttpClient("PnyxWebAssembly.ServerAPI.Private", client => client.BaseAddress = new Uri(builder.HostEnvironment.BaseAddress))
                .AddHttpMessageHandler<BaseAddressAuthorizationMessageHandler>();

            builder.Services.AddHttpClient("PnyxWebAssembly.ServerAPI.Public", client => client.BaseAddress = new Uri(builder.HostEnvironment.BaseAddress));

            builder.Services.AddSingleton<IssueImageService>();

            builder.Services.AddSingleton<ImageCacheService>();

            builder.Services.AddSingleton<UserCacheService>();

            builder.Services.AddSingleton<UserService>();

            builder.Services.AddApiAuthorization();

            builder.Services.AddMudServices();

            await builder.Build().RunAsync();
        }
    }
}
