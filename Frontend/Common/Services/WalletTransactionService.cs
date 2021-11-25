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
    /// Implementation of the wallet transaction service
    /// </summary>
    public class WalletTransactionService
    {
        /// <summary>
        /// Gets the wallet transactions for user.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="userId">The user identifier.</param>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="limit">The limit.</param>
        /// <returns>
        /// The wallet transactions for the use
        /// </returns>
        public List<WalletTransaction> GetWalletTransactionsForUser(DbServiceContext dbServiceContext, string userId, PaginatedList paginatedList, int limit = 0)
        {
            User user = dbServiceContext.Users
                .Include(u => u.Wallet)
                .Include(u => u.Wallet.WalletTransactions)
                .FirstOrDefault(u => u.Id.ToString() == userId);

            if (user == null)
            {
                return null;
            }

            List<WalletTransaction> walletTransactions;

            if (limit == 0)
            {
                walletTransactions = user.Wallet.WalletTransactions
                    .OrderByDescending(w => w.CreateDate)
                    .ToList();
            }
            else
            {
                walletTransactions = user.Wallet.WalletTransactions
                    .OrderByDescending(w => w.CreateDate)
                    .Take(limit)
                    .ToList();
            }

            Dictionary<string, TransactionType> transactionTypes = dbServiceContext.TransactionTypes
                .ToDictionary(t => t.Id.ToString(), t => t);

            foreach (WalletTransaction walletTransaction in walletTransactions)
            {
                walletTransaction.TransactionType = transactionTypes[walletTransaction.TransactionTypeId.ToString()];
            }

            return ProcessPaginatedList(paginatedList, walletTransactions);
        }

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>The number of imported records</returns>
        public int Import(DataTable dataTable)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                int count = dbServiceContext.WalletTransactions.Count();

                if (count > 0)
                {
                    return 0;
                }
                
                Dictionary<string, TransactionType> transactionTypes = dbServiceContext.TransactionTypes
                    .Where(t => t.ImportId != null)
                    .ToDictionary(t => t.ImportId.ToString(), t => t);

                int recordCount = 0;

                foreach (DataRow row in dataTable.Rows)
                {
                    int id = Convert.ToInt32(row["WalletID"].ToString());

                    User user = dbServiceContext.Users
                        .Include(u => u.Wallet)
                        .Include(u => u.Wallet.WalletTransactions)
                        .FirstOrDefault(u => u.Wallet.ImportId == id);

                    if (user != null && user.Wallet != null)
                    {
                        string key = row["TransactionTypeID"].ToString();

                        if (!string.IsNullOrEmpty(key) && transactionTypes.ContainsKey(key))
                        {
                            TransactionType transactionType = transactionTypes[key];

                            WalletTransaction walletTransaction = new WalletTransaction
                            {
                                ImportId = Convert.ToInt32(row["ID"].ToString()),
                                Balance = Convert.ToDouble(row["Balance"].ToString()),
                                TransactionType = transactionType,
                                WalletId = user.Wallet.Id
                            };

                            string importId = row["TransactionID"].ToString();

                            if (!string.IsNullOrEmpty(importId))
                            {
                                if (transactionType.Name == TransactionTypeNames.StakeProposal.ToString())
                                {
                                    Guid? transactionId = dbServiceContext.StakedProposals
                                        .FirstOrDefault(s => s.ImportId == Convert.ToInt32(importId))?.Id;

                                    walletTransaction.TransactionId = transactionId;
                                }
                            }

                            dbServiceContext.WalletTransactions.Add(walletTransaction);

                            recordCount++;
                        }
                    }
                }

                if (recordCount > 0)
                {
                    dbServiceContext.SaveChanges();
                }

                return recordCount;
            }
        }

        /// <summary>
        /// Processes the paginated list.
        /// </summary>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="walletTransactions">The wallet transactions.</param>
        /// <returns></returns>
        private static List<WalletTransaction> ProcessPaginatedList(PaginatedList paginatedList, List<WalletTransaction> walletTransactions)
        {
            if (paginatedList is { ItemsPerPage: > 0 })
            {
                walletTransactions = walletTransactions
                    .Skip(paginatedList.Skip)
                    .Take(paginatedList.ItemsPerPage)
                    .ToList();
            }

            return walletTransactions;
        }
    }
}
