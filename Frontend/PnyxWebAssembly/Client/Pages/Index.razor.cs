using Common.Entities;
using Microsoft.AspNetCore.Components;
using Microsoft.AspNetCore.Components.Authorization;
using Microsoft.JSInterop;
using PnyxWebAssembly.Client.Components;
using PnyxWebAssembly.Client.Services;
using PnyxWebAssembly.Client.Shared;
using System;
using System.Collections.Generic;
using System.Globalization;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Net.Http.Json;
using System.Security.Claims;
using System.Threading.Tasks;

namespace PnyxWebAssembly.Client.Pages
{
    /// <summary>
    /// Implementation of the index code behind class
    /// </summary>
    /// <seealso cref="Microsoft.AspNetCore.Components.ComponentBase" />
    public partial class Index
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

        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        [Inject]
        private IHttpClientFactory ClientFactory { get; set; }

        /// <summary>
        /// Gets or sets the image cache service.
        /// </summary>
        /// <value>
        /// The image cache service.
        /// </value>
        [Inject]
        private ImageCacheService ImageCacheService { get; set; }

        /// <summary>
        /// Gets or sets the main layout.
        /// </summary>
        /// <value>
        /// The main layout.
        /// </value>
        [CascadingParameter]
        public MainLayout MainLayout { get; set; }

        /// <summary>
        /// Gets or sets the user add wizard.
        /// </summary>
        /// <value>
        /// The user add wizard.
        /// </value>
        [CascadingParameter]
        public UserAddWizard UserAddWizard { get; set; }

        /// <summary>
        /// Gets or sets the issue card.
        /// </summary>
        /// <value>
        /// The issue card.
        /// </value>
        [CascadingParameter]
        public IssueCard IssueCard { get; set; }

        /// <summary>
        /// Gets or sets the name of the user.
        /// </summary>
        /// <value>
        /// The name of the user.
        /// </value>
        public string UserEMail { get; set; }

        /// <summary>
        /// Gets or sets the name of the user.
        /// </summary>
        /// <value>
        /// The name of the user.
        /// </value>
        public string UserName { get; set; }

        /// <summary>
        /// Gets or sets the height.
        /// </summary>
        /// <value>
        /// The height.
        /// </value>
        public int Height { get; set; }

        /// <summary>
        /// Gets or sets the width.
        /// </summary>
        /// <value>
        /// The width.
        /// </value>
        public int Width { get; set; }

        /// <summary>
        /// Gets or sets the cards per row.
        /// </summary>
        /// <value>
        /// The cards per row.
        /// </value>
        public int CardsPerRow { get; set; }

        /// <summary>
        /// Gets or sets the page.
        /// </summary>
        /// <value>
        /// The page.
        /// </value>
        public int Page { get; set; } = 1;

        /// <summary>
        /// Gets or sets the issue.
        /// </summary>
        /// <value>
        /// The issue.
        /// </value>
        public List<Issue> Issues { get; set; }

        /// <summary>
        /// Gets or sets the external user identifier.
        /// </summary>
        /// <value>
        /// The external user identifier.
        /// </value>
        public Guid ExternalUserId { get; set; }

        /// <summary>
        /// Gets or sets a value indicating whether [show add user wizard].
        /// </summary>
        /// <value>
        ///   <c>true</c> if [show add user wizard]; otherwise, <c>false</c>.
        /// </value>
        public bool ShowAddUserWizard { get; set; }

        /// <summary>
        /// Method invoked when the component is ready to start, having received its
        /// initial parameters from its parent in the render tree.
        /// Override this method if you will perform an asynchronous operation and
        /// want the component to refresh when that operation is completed.
        /// </summary>
        protected override async Task OnInitializedAsync()
        {
            AvatarImageService.ClientFactory = ClientFactory;

            await ManageWindowResizing();

            await UpdateUserInfo();
        }

        /// <summary>
        /// Updates the user information.
        /// </summary>
        private async Task UpdateUserInfo()
        {
            ShowAddUserWizard = false;
            UserService.User = null;
            MainLayout.UserName = string.Empty;
            MainLayout.TotalBalance = -1;

            AuthenticationState authState = await AuthenticationStateProvider.GetAuthenticationStateAsync();
            ClaimsPrincipal user = authState.User;

            UserEMail = user?.Identity?.Name;

            if (user != null && !string.IsNullOrEmpty(UserEMail))
            {
                List<Claim> claims = user.Claims.ToList();

                if (claims.Count >= 3)
                {
                    ExternalUserId = Guid.Parse(claims[2].Value);
                }

                using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Private");

                await LogService.LogToServer(client, $"Login {UserEMail} {ExternalUserId}");

                User userFromService;

                try
                {
                    // TODO: remove later - external user id is to request for
                    // this is just for presentation issues for docker
                    if (UserEMail != "andreasschurz@andreasschurz.de")
                    {
                        userFromService = await client.GetFromJsonAsync<User>($"User/ByExternalId/{ExternalUserId}");
                    }
                    else
                    {
                        userFromService = await client.GetFromJsonAsync<User>($"User/ByName/Andreas");
                    }
                }
                catch (HttpRequestException ex)
                {
                    if (ex.StatusCode == HttpStatusCode.NotFound)
                    {
                        userFromService = null;
                    }
                    else
                    {
                        throw;
                    }
                }

                if (userFromService == null)
                {
                    ShowAddUserWizard = true;
                    return;
                }

                double totalBalance = userFromService.Wallet.TotalBalance;

                UserService.User = userFromService;

                UserName = userFromService.UserName;
                MainLayout.UserName = userFromService.UserName;
                MainLayout.TotalBalance = int.Parse(Math.Round(totalBalance, 0).ToString(CultureInfo.InvariantCulture));

                string avatarImage;

                if (!ImageCacheService.HasImage(userFromService.UserName))
                {
                    avatarImage = await AvatarImageService.GetAvatarImageBase64(userFromService.UserName);
                    ImageCacheService.Add(userFromService.UserName, avatarImage);
                }
                else
                {
                    avatarImage = ImageCacheService.Get(userFromService.UserName);
                }

                MainLayout.AvatarImage = avatarImage;

                if (CardsPerRow == 0)
                {
                    CardsPerRow = 4;
                }

                Issues ??= new List<Issue>();

                await AddIssues(client);
            }
        }

        /// <summary>
        /// Adds the issues.
        /// </summary>
        /// <param name="client">The client.</param>
        private async Task AddIssues(HttpClient client)
        {
            int itemsPerPage = 4;

            try
            {
                List<Issue> issues = await client.GetFromJsonAsync<List<Issue>>(
                    $"Issues?UserName={UserName}&Page={Page}&ItemsPerPage={itemsPerPage}");

                if (issues != null)
                {
                    Issues.AddRange(issues);
                }
            }
            catch (HttpRequestException ex)
            {
                if (ex.StatusCode != HttpStatusCode.NotFound)
                {
                    throw;
                }
            }
        }

        /// <summary>
        /// Manages the window resizing.
        /// </summary>
        private async Task ManageWindowResizing()
        {
            BrowserResizeService.JsRuntime = JsRuntime;

            BrowserResizeService.OnResize += BrowserHasResized;

            await JsRuntime.InvokeAsync<object>("browserResize.registerResizeCallback");

            await GetDimensions();
        }

        /// <summary>
        /// Browsers the has resized.
        /// </summary>
        private async Task BrowserHasResized()
        {
            await GetDimensions();
        }

        /// <summary>
        /// Gets the dimensions.
        /// </summary>
        private async Task GetDimensions()
        {
            Height = await BrowserResizeService.GetInnerHeight() - 85;
            Width = await BrowserResizeService.GetInnerWidth();

            double cardsPerRow = Math.Floor(Width / 330D);

            if (cardsPerRow > 4)
            {
                cardsPerRow = 4;
            }

            CardsPerRow = Convert.ToInt32(cardsPerRow);
        }

        /// <summary>
        /// Called when [success].
        /// </summary>
        private async Task OnSuccess()
        {
            await UpdateUserInfo();

            await InvokeAsync(StateHasChanged);
        }

        private async Task UpdateIssues()
        {
            Page++;

            using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");

            await AddIssues(client);
        }
    }
}
