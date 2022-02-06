using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Security.Claims;
using System.Threading.Tasks;
using Common.Entities;
using Microsoft.AspNetCore.Components;
using Microsoft.AspNetCore.Components.Authorization;
using MudBlazor;
using PnyxWebAssembly.Client.Services;

namespace PnyxWebAssembly.Client.Components
{
    public partial class IssueEditor
    {
        /// <summary>
        /// Gets or sets the navigation manager.
        /// </summary>
        /// <value>
        /// The navigation manager.
        /// </value>
        [Inject]
        private NavigationManager NavigationManager { get; set; }

        /// <summary>
        /// Gets or sets the authentication state provider.
        /// </summary>
        /// <value>
        /// The authentication state provider.
        /// </value>
        [Inject]
        private AuthenticationStateProvider AuthenticationStateProvider { get; set; }

        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        [Inject]
        private IHttpClientFactory ClientFactory { get; set; }

        [Parameter]
        public string Action { get; set; }

        public IssueService IssueService { get; set; }

        private Issue Issue { get; set; }

        private EditMode EditMode { get; set; }

        private User User { get; set; }

        private string ErrorMessage { get; set; }

        private MudForm Form { get; set; }

        private bool ShowHashtagPopover { get; set; }

        private string HashtagValue { get; set; }

        protected override async Task OnInitializedAsync()
        {
            Issue = new Issue
            {
                Tags = string.Empty
            };

            IssueService = new IssueService();
            IssueService.ClientFactory = ClientFactory;

            AuthenticationState authState = await AuthenticationStateProvider.GetAuthenticationStateAsync();
            ClaimsPrincipal user = authState.User;

            string userEMail = user?.Identity?.Name;

            if (user != null && !string.IsNullOrEmpty(userEMail))
            {
                List<Claim> claims = user.Claims.ToList();

                if (claims.Count >= 3)
                {
                    string externalUserId = claims[2].Value;

                    if (UserService.User != null && 
                        UserService.User.UniqueExternalUserId.ToString() == externalUserId)
                    {
                        User = UserService.User;
                    }
                }
            }

            if (Action.Equals("add", StringComparison.OrdinalIgnoreCase))
            {
                if (User != null)
                {
                    EditMode = EditMode.AddNew;
                }
                else
                {
                    ErrorMessage = "Error: user has no permission to add issue";
                    return;
                }
            }
            else
            {
                if (!Guid.TryParse(Action, out Guid issueId))
                {
                    ErrorMessage = "Error: given parameter is no valid issue id";
                    return;
                }

                HttpClient client;

                if (User != null)
                {
                    client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Private");
                    EditMode = EditMode.Edit;
                }
                else
                {
                    client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");
                    EditMode = EditMode.ReadOnly;
                }
                
                // TODO: load issue with all proposals
            }
        }

        private void AddHashtagClick()
        {
            ShowHashtagPopover = true;
        }

        private void OnMudChipClose(MudChip chip)
        {
            throw new NotImplementedException();
        }

        private void CloseHashtagPopover()
        {
            ShowHashtagPopover = false;
        }

        private Task<IEnumerable<string>> SearchHashtags(string value)
        {
            return IssueService.GetHashtags(value);
        }
    }
}
