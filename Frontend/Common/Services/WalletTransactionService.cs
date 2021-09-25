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
        /// <param name="fromDate">From date.</param>
        /// <param name="toDate">To date.</param>
        /// <returns>The wallet transactions for the use</returns>
        public List<WalletTransaction> GetWalletTransactionsForUser(DbServiceContext dbServiceContext, Guid userId, DateTime? fromDate,
            DateTime? toDate)
        {
            User user = dbServiceContext.User
                .Include(u => u.Wallet)
                .Include(u => u.Wallet.WalletTransactions)
                .FirstOrDefault(u => u.Id.ToString() == userId.ToString());

            if (user == null)
            {
                return null;
            }

            if (fromDate == null && toDate == null)
            {
                return user.Wallet.WalletTransactions;
            }

            if (fromDate == null)
            {
                return user.Wallet.WalletTransactions
                    .Where(w => w.CreateDate <= (DateTime) toDate).ToList();
            }

            if (toDate == null)
            {
                return user.Wallet.WalletTransactions
                    .Where(w => w.CreateDate >= (DateTime) fromDate).ToList();
            }

            return user.Wallet.WalletTransactions
                .Where(w => w.CreateDate <= (DateTime) toDate)
                .Where(w => w.CreateDate >= (DateTime) fromDate).ToList();
        }

        /// <summary>
        /// Adds the wallet transaction.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="walletTransaction">The wallet transaction.</param>
        public void AddWalletTransaction(DbServiceContext dbServiceContext, WalletTransaction walletTransaction)
        {
            dbServiceContext.WalletTransactions.Add(walletTransaction);
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

                    User user = dbServiceContext.User
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
                                if (transactionType.Name == TransactionTypeNames.StakeSuggestion.ToString())
                                {
                                    Guid? transactionId = dbServiceContext.StakedSuggestions
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
    }
}
