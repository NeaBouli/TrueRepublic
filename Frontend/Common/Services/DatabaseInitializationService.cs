using System;
using Common.Data;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the DatabaseInitializationService
    /// </summary>
    public static class DatabaseInitializationService
    {
        /// <summary>
        /// Gets or sets the database connect string.
        /// </summary>
        /// <value>
        /// The database connect string.
        /// </value>
        public static string DbConnectString { get; set; }

        /// <summary>
        /// Gets or sets the database authentication connect string.
        /// </summary>
        /// <value>
        /// The database authentication connect string.
        /// </value>
        public static string DbAuthConnectString { get; set; }

        /// <summary>
        /// Gets or sets a value indicating whether this instance is docker.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance is docker; otherwise, <c>false</c>.
        /// </value>
        public static bool IsDocker { get; set; }

        /// <summary>
        /// Creates the database if not existing.
        /// </summary>
        /// <exception cref="System.InvalidOperationException">Will be thrown if the connect string is empty</exception>
        public static void CreateDbIfNotExisting()
        {
            DbServiceContext dbServiceContext = GetDbServiceContext();

            using (dbServiceContext)
            {
                dbServiceContext.Database.EnsureCreated();
            }
        }

        /// <summary>
        /// Gets the database service context.
        /// </summary>
        /// <returns>The Db Service context</returns>
        /// <exception cref="System.InvalidOperationException">Will be thrown if DbConnectString is not set</exception>
        public static DbServiceContext GetDbServiceContext()
        {
            if (string.IsNullOrEmpty(DbConnectString))
            {
                throw new InvalidOperationException(Resource.ErrorDbConnectStringCannotBeEmpty);
            }

            return new DbServiceContext(DbConnectString);
        }
    }
}
