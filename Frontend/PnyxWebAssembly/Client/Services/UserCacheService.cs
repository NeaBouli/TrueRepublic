using Common.Entities;

namespace PnyxWebAssembly.Client.Services
{
    /// <summary>
    /// Implementation of the user cache service
    /// </summary>
    public static class UserCacheService
    {
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
