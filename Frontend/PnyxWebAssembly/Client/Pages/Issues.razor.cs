using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Json;
using System.Security.Claims;
using System.Threading.Tasks;
using Common.Entities;
using Microsoft.AspNetCore.Components;
using Microsoft.AspNetCore.Components.Authorization;

namespace PnyxWebAssembly.Client.Pages
{
    public partial class Issues
    {
        [Inject]
        private IHttpClientFactory ClientFactory { get; set; }

        [Inject]
        private AuthenticationStateProvider AuthenticationStateProvider { get; set; }

        private List<Issue> IssueItems { get; set; }

        private string UserName { get; set; }

        protected override async Task OnInitializedAsync()
        {
            AuthenticationState authState = await AuthenticationStateProvider.GetAuthenticationStateAsync();
            ClaimsPrincipal user = authState.User;

            UserName = user?.Identity?.Name;

            if (string.IsNullOrEmpty(UserName))
            {
                UserName = "Unknown";
            }

            HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");

            IssueItems = await client.GetFromJsonAsync<List<Issue>>("Issues");
        }
    }
}
