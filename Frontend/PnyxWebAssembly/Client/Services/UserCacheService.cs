using System.Collections.Concurrent;
using Common.Entities;

namespace PnyxWebAssembly.Client.Services
{
    public class UserCacheService
    {
        /// <summary>
        /// The image cache
        /// </summary>
        private readonly ConcurrentDictionary<string, User> _userCache = new();

        /// <summary>
        /// Adds the specified name.
        /// </summary>
        /// <param name="name">The name.</param>
        /// <param name="user">The user.</param>
        public void Add(string name, User user)
        {
            _userCache.TryAdd(name, user);
        }

        /// <summary>
        /// Determines whether the specified name has user.
        /// </summary>
        /// <param name="name">The name.</param>
        /// <returns>
        ///   <c>true</c> if the specified name has user; otherwise, <c>false</c>.
        /// </returns>
        public bool HasUser(string name)
        {
            return _userCache.ContainsKey(name);
        }

        /// <summary>
        /// Gets the specified name.
        /// </summary>
        /// <param name="name">The name.</param>
        /// <returns>The user</returns>
        public User Get(string name)
        {
            _userCache.TryGetValue(name, out User value);

            return value;
        }
    }
}
