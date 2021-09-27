using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Data;
using Common.Entities;
using Microsoft.EntityFrameworkCore;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the user service
    /// </summary>
    public class UserService
    {
        /// <summary>
        /// Gets the user by identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="id">The identifier.</param>
        /// <returns>
        /// The user if found otherwise null
        /// </returns>
        public User GetUserById(DbServiceContext dbServiceContext, Guid id)
        {
            return dbServiceContext.User
                .Include(u => u.Wallet)
                .FirstOrDefault(u => u.Id.ToString() == id.ToString());
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
            using DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();
            
            if (dbServiceContext.User.ToList().Count > 0)
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

                if (dataTable.Columns.Contains("UniqueExternalUserId"))
                {
                    var uniqueExternalUserId = row["UniqueExternalUserId"].ToString();

                    if (!string.IsNullOrEmpty(uniqueExternalUserId))
                    {
                        user.UniqueExternalUserId = Guid.Parse(uniqueExternalUserId);
                    }
                    else
                    {
                        user.UniqueExternalUserId = Guid.NewGuid();
                    }
                }
                else
                {
                    user.UniqueExternalUserId = Guid.NewGuid();
                }

                dbServiceContext.User.Add(user);

                recordCount++;
            }

            if (recordCount > 0)
            {
                dbServiceContext.SaveChanges();
            }

            return recordCount;
        }
    }
}
