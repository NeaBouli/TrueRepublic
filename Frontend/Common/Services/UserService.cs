using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Data;
using Common.Entities;
using Common.Interfaces;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the user service
    /// </summary>
    public class UserService : IUserService
    {
        /// <summary>
        /// Gets the users.
        /// </summary>
        /// <returns></returns>
        public List<User> GetUsers()
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                return dbServiceContext.User.ToList();
            }
        }

        /// <summary>
        /// Gets the user by identifier.
        /// </summary>
        /// <param name="id">The identifier.</param>
        /// <returns>The user if found otherwise null</returns>
        public User GetUserById(Guid id)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                return dbServiceContext.User.FirstOrDefault(u => u.Id.ToString() == id.ToString());
            }
        }

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>
        /// The number of imported records
        /// </returns>
        public int Import(DataTable dataTable)
        {
            List<User> users = GetUsers();

            if (users.Count > 0)
            {
                return 0;
            }

            int recordCount = 0;

            foreach (DataRow row in dataTable.Rows)
            {
                User user = new()
                {
                    UserName = row["UserName"].ToString(),
                    ImportId = Convert.ToInt32(row["ID"].ToString()),
                    StakedSuggestions = new List<StakedSuggestion>()
                };

                string uniqueExternalUserId = row["UniqueExternalUserId"].ToString();
                user.UniqueExternalUserId = string.IsNullOrEmpty(uniqueExternalUserId) ?
                    Guid.NewGuid() : Guid.Parse(uniqueExternalUserId);

                Add(user);

                recordCount++;
            }

            return recordCount;
        }

        /// <summary>
        /// Adds the specified user.
        /// </summary>
        /// <param name="user">The user.</param>
        public void Add(User user)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                dbServiceContext.User.Add(user);
                dbServiceContext.SaveChanges();
            }
        }
    }
}
