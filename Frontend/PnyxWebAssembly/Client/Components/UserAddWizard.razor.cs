using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Net.Http.Json;
using System.Threading.Tasks;
using Common.Entities;
using Microsoft.AspNetCore.Components;
using MudBlazor;

namespace PnyxWebAssembly.Client.Components
{
    /// <summary>
    /// Implementation of the user add wizard
    /// </summary>
    /// <seealso cref="Microsoft.AspNetCore.Components.ComponentBase" />
    public partial class UserAddWizard
    {
        /// <summary>
        /// Gets or sets the external user identifier.
        /// </summary>
        /// <value>
        /// The external user identifier.
        /// </value>
        [Parameter]
        public Guid ExternalUserId { get; set; }

        /// <summary>
        /// Gets or sets the username.
        /// </summary>
        /// <value>
        /// The username.
        /// </value>
        public string UserName { get; set; }

        [Parameter]
        public EventCallback OnSuccess { get; set; }

        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        [Inject]
        private IHttpClientFactory ClientFactory { get; set; }

        private bool _isRunning;

        /// <summary>
        /// The success
        /// </summary>
        private bool _success;

        /// <summary>
        /// The errors
        /// </summary>
        private string[] _errors = { };

        /// <summary>
        /// The form
        /// </summary>
        private MudForm _form;

        /// <summary>
        /// Validates the name of the user.
        /// </summary>
        /// <param name="userName">Name of the user.</param>
        /// <returns></returns>
        private IEnumerable<string> ValidateUserName(string userName)
        {
            if (string.IsNullOrEmpty(userName))
            {
                yield return "Username is required";
                yield break;
            }

            if (userName.Length < 5)
            {
                yield return "Username must at least be 5 characters long";
            }
            else if (userName.Length > 15)
            {
                yield return "Username can be maximum 15 characters long";
            }

            if (userName.Contains(" "))
            {
                yield return "Username must not contain whitespaces";
            }

            string invalidChars = "?&^$#@!()+-,:;<>’\'-_*";

            foreach (char c in invalidChars)
            {
                if (userName.Contains(c.ToString()))
                {
                    yield return $"Username must not contain the following characters \"{invalidChars}\" - Found: \"{c}\"";
                    yield break;
                }
            }
        }

        /// <summary>
        /// Checks the user name already existing.
        /// </summary>
        private async void CreateUserIfNotAlreadyExisting()
        {
            if (ValidateUserName(UserName).Any())
            {
                return;
            }

            if (_isRunning)
            {
                return;
            }

            _isRunning = true;

            try
            {
                await CreateUserIfNotAlreadyExistingSave();
            }
            finally
            {
                _isRunning = false;
            }
        }

        private async Task CreateUserIfNotAlreadyExistingSave()
        {
            using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Private");

            User userFromService;

            try
            {
                userFromService = await client.GetFromJsonAsync<User>($"User/ByName/{UserName}");
            }
            catch (HttpRequestException)
            {
                userFromService = null;
            }

            if (userFromService != null)
            {
                _errors = new[] {$"User \"{UserName}\" is already existing. Please select a different name"};
                await InvokeAsync(StateHasChanged);
                return;
            }

            User user = new User
            {
                UserName = UserName,
                UniqueExternalUserId = ExternalUserId
            };

            using var response = await client.PostAsJsonAsync("User", user);

            if (response.IsSuccessStatusCode)
            {
                _success = true;

                await InvokeAsync(StateHasChanged);

                await OnSuccess.InvokeAsync();
            }
        }
    }
}
