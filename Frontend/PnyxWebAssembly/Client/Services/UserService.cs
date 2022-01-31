using System;
using System.Net.Http;
using System.Net.Http.Json;
using System.Threading.Tasks;
using Common.Entities;

namespace PnyxWebAssembly.Client.Services
{
    /// <summary>
    /// Implementation of the user cache service
    /// </summary>
    public static class UserService
    {
        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        public static IHttpClientFactory ClientFactory { get; set; }

        /// <summary>
        /// Gets or sets the user cache service.
        /// </summary>
        /// <value>
        /// The user cache service.
        /// </value>
        public static UserCacheService UserCacheService { get; set; }

        /// <summary>
        /// Gets the user by identifier.
        /// </summary>
        /// <param name="id">The identifier.</param>
        /// <returns>The user for the given id</returns>
        public static async Task<User> GetUserById(Guid id)
        {
            User user;

            if (!IsAuthenticated)
            {
                user = new User
                {
                    UserName = "Unknown: Login for details",
                    Id = Guid.Empty
                };

                return user;
            }

            using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Private");

            if (UserCacheService.HasUser(id.ToString()))
            {
                user = UserCacheService.Get(id.ToString());

                await LogService.LogToServer(client, $"User {user.Id} ({user.UserName}) taken from user cache");

                return user;
            }

            user = await client.GetFromJsonAsync<User>($"User/ById/{id}");

            if (user != null && !UserCacheService.HasUser(user.Id.ToString()))
            {
                UserCacheService.Add(id.ToString(), user);
            }

            return user;
        }

        /// <summary>
        /// Gets or sets the user.
        /// </summary>
        /// <value>
        /// The user.
        /// </value>
        public static User User { get; set; }

        /// <summary>
        /// Gets a value indicating whether this instance is authenticated.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance is authenticated; otherwise, <c>false</c>.
        /// </value>
        public static bool IsAuthenticated => User != null;
    }
}
