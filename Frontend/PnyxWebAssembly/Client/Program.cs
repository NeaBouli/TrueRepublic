using Microsoft.AspNetCore.Components.WebAssembly.Authentication;
using Microsoft.AspNetCore.Components.WebAssembly.Hosting;
using Microsoft.Extensions.DependencyInjection;
using System;
using System.Net.Http;
using System.Threading.Tasks;

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

            // TODO: might be removed later
            // Supply HttpClient instances that include access tokens when making requests to the server project
            // builder.Services.AddScoped(sp => sp.GetRequiredService<IHttpClientFactory>().CreateClient("PnyxWebAssembly.ServerAPI.Private"));

            builder.Services.AddHttpClient("PnyxWebAssembly.ServerAPI.Public", client => client.BaseAddress = new Uri(builder.HostEnvironment.BaseAddress));

            // TODO: might be removed later
            // builder.Services.AddScoped(sp => sp.GetRequiredService<IHttpClientFactory>().CreateClient("PnyxWebAssembly.ServerAPI.Public"));

            builder.Services.AddApiAuthorization();

            await builder.Build().RunAsync();
        }
    }
}
