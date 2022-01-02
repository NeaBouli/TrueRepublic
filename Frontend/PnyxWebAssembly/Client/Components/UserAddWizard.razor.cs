using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Json;
using Common.Entities;

namespace PnyxWebAssembly.Client.Components
{
    /// <summary>
    /// Implementation of the user add wizard
    /// </summary>
    /// <seealso cref="Microsoft.AspNetCore.Components.ComponentBase" />
    public partial class UserAddWizard
    {
        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        public IHttpClientFactory ClientFactory { get; set; }

        /// <summary>
        /// Validates the name of the user.
        /// </summary>
        /// <param name="userName">Name of the user.</param>
        /// <returns></returns>
        private async IAsyncEnumerable<string> ValidateUserNameAsync(string userName)
        {
            if (string.IsNullOrWhiteSpace(userName))
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

            using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Private");

            User userFromService = await client.GetFromJsonAsync<User>($"User/ByName/{userName}");

            if (userFromService != null)
            {
                yield return "A user with the same name already exists. Please select a different name";
            }
        }
    }
}
