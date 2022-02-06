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

        /// <summary>
        /// Semaphore to prevent multiple clicking
        /// </summary>
        private bool _isRunning;

        /// <summary>
        /// The success flag
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
        /// <returns>The user name</returns>
        private IEnumerable<string> ValidateUserName(string userName)
        {
            if (string.IsNullOrEmpty(userName))
            {
                yield return "Ein Benutzername muss angegeben werden";
                yield break;
            }

            if (userName.Length < 5)
            {
                yield return "Der Benutzername muss mindestens 5 Zeichen lang sein";
            }
            else if (userName.Length > 15)
            {
                yield return "Der Benutzername kann maximal 15 Zeichen lang sein";
            }

            if (userName.Contains(" "))
            {
                yield return "Der Benutzername darf keine Leerzeichen enthalten";
            }

            string invalidChars = "?&^$#@!()+-,:;<>’\'-_*";

            foreach (char c in invalidChars)
            {
                if (userName.Contains(c.ToString()))
                {
                    yield return $"Der Benutzername darf die folgenden Zeichen nicht enthalten \"{invalidChars}\" - Gefunden: \"{c}\"";
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

        /// <summary>
        /// Creates the user if not already existing save.
        /// </summary>
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
                _errors = new[] {$"Der Benutzer \"{UserName}\" existiert bereits. Bitte wählen Sie einen anderen Namen"};
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
