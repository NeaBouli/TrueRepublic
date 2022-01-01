using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Json;
using System.Security.Claims;
using System.Threading.Tasks;
using Common.Entities;
using Common.GuiEntities;
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

        private List<RenderIssue> IssueItems { get; set; }

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

            List<Issue> issueItems;

            using (HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public"))
            {
                issueItems = await client.GetFromJsonAsync<List<Issue>>("Issues");
            }

            IssueItems = new List<RenderIssue>();

            if (issueItems != null)
            {
                foreach (var issueItem in issueItems)
                {
                    IssueItems.Add(new RenderIssue(issueItem));
                }
            }
        }
    }
}
