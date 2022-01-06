using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Json;
using System.Threading.Tasks;
using Common.Entities;

namespace PnyxWebAssembly.Client.Services
{
    /// <summary>
    /// Implementation of the user cache service
    /// </summary>
    public static class UserCacheService
    {
        /// <summary>
        /// The user
        /// </summary>
        private static User _user;

        /// <summary>
        /// The users by identifier
        /// </summary>
        private static readonly Dictionary<string, User> UsersById = new();

        /// <summary>
        /// The users by name
        /// </summary>
        private static readonly Dictionary<string, User> UsersByName = new();

        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        public static IHttpClientFactory ClientFactory { get; set; }

        /// <summary>
        /// Gets the user by identifier.
        /// </summary>
        /// <param name="id">The identifier.</param>
        /// <returns>The user for the given id</returns>
        public static async Task<User> GetUserById(Guid id)
        {
            if (!UsersById.ContainsKey(id.ToString()))
            {
                using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Private");

                User user = await client.GetFromJsonAsync<User>($"User/ById/{id}");

                if (user != null)
                {
                    UsersById.Add(user.Id.ToString(), user);

                    if (!UsersByName.ContainsKey(user.UserName))
                    {
                        UsersByName.Add(user.UserName, user);
                    }
                }

                return user;
            }

            return UsersById[id.ToString()];
        }

        /// <summary>
        /// Gets the name of the user by.
        /// </summary>
        /// <param name="userName">Name of the user.</param>
        /// <returns>The user for the given name</returns>
        public static async Task<User> GetUserByName(string userName)
        {
            if (!UsersByName.ContainsKey(userName))
            {
                using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Private");

                User user = await client.GetFromJsonAsync<User>($"User/ByName/{userName}");

                if (user != null)
                {
                    UsersByName.Add(userName, user);

                    if (!UsersById.ContainsKey(user.Id.ToString()))
                    {
                        UsersById.Add(user.Id.ToString(), user);
                    }
                }
                
                return user;
            }

            return UsersByName[userName];
        }

        /// <summary>
        /// Gets or sets the user.
        /// </summary>
        /// <value>
        /// The user.
        /// </value>
        public static User User
        {
            get => _user;
            set
            {
                _user = value;

                if (_user == null)
                {
                    return;
                }

                if (!UsersById.ContainsKey(_user.Id.ToString()))
                {
                    UsersById.Add(_user.Id.ToString(), _user);
                }

                if (!UsersByName.ContainsKey(_user.UserName))
                {
                    UsersByName.Add(_user.UserName, _user);
                }
            }
        }

        /// <summary>
        /// Gets a value indicating whether this instance is authenticated.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance is authenticated; otherwise, <c>false</c>.
        /// </value>
        public static bool IsAuthenticated => User != null;
    }
}
