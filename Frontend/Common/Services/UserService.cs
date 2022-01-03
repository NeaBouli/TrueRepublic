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
        /// Gets the name of the user by.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="userName">Name of the user.</param>
        /// <returns>The user</returns>
        public User GetUserByName(DbServiceContext dbServiceContext, string userName)
        {
            return dbServiceContext.Users
                .Include(u => u.Wallet)
                .FirstOrDefault(u => u.UserName == userName);
        }

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
            return dbServiceContext.Users
                .Include(u => u.Wallet)
                .FirstOrDefault(u => u.Id.ToString() == id.ToString());
        }

        /// <summary>
        /// Gets the user by external identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="id">The identifier.</param>
        /// <returns>
        /// The user if found otherwise null
        /// </returns>
        public User GetUserByExternalId(DbServiceContext dbServiceContext, Guid id)
        {
            return dbServiceContext.Users
                .Include(u => u.Wallet)
                .FirstOrDefault(u => u.UniqueExternalUserId.ToString() == id.ToString());
        }

        /// <summary>
        /// Gets the user identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="userName">Name of the user.</param>
        /// <returns>The user id as string if found otherwise an empty string</returns>
        public string GetUserId(DbServiceContext dbServiceContext, string userName)
        {
            string userId = string.Empty;

            if (!string.IsNullOrEmpty(userName))
            {
                User user = GetUserByName(dbServiceContext, userName);

                if (user != null)
                {
                    userId = user.Id.ToString();
                }
            }

            return userId;
        }

        /// <summary>
        /// Creates the user.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="user">The user.</param>
        /// <param name="genesisFunding">The genesis funding.</param>
        public void CreateUser(DbServiceContext dbServiceContext, User user, double genesisFunding = 0)
        {
            Wallet wallet = new Wallet
            {
                TotalBalance = 0
            };

            user.WalletId = wallet.Id;

            dbServiceContext.Wallets.Add(wallet);

            dbServiceContext.Users.Add(user);

            dbServiceContext.SaveChanges();

            if (genesisFunding == 0)
            {
                return;
            }

            WalletService walletService = new WalletService();

            walletService.AddTransaction(dbServiceContext, user.Id, TransactionTypeNames.Genesis, null, genesisFunding);

            dbServiceContext.SaveChanges();
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
            
            if (dbServiceContext.Users.ToList().Count > 0)
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
                    StakedSuggestions = new List<StakedProposal>()
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

                dbServiceContext.Users.Add(user);

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
