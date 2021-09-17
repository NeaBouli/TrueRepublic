using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Data;
using Common.Entities;
using Common.Interfaces;
using Microsoft.EntityFrameworkCore;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the wallet service
    /// </summary>
    public class WalletService : IWalletService
    {
        /// <summary>
        /// Gets the total balance ot all wallets
        /// </summary>
        /// <returns>The total balance of all wallets</returns>
        public double GetTotalBalance()
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            double totalBalance = 0;

            using (dbServiceContext)
            {
                return Enumerable.Aggregate(dbServiceContext.Wallets, totalBalance, (current, wallet) => current + wallet.TotalBalance);
            }
        }

        /// <summary>
        /// Determines whether [has enough funding for transaction] [the specified user identifier].
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <returns>
        ///   <c>true</c> if [has enough funding for transaction] [the specified user identifier]; otherwise, <c>false</c>.
        /// </returns>
        public bool HasEnoughFundingForTransaction(Guid userId, TransactionTypeNames transactionTypeName)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                TransactionType transactionType = GetTransActionType(dbServiceContext, transactionTypeName);

                Wallet wallet = GetWalletForUserId(dbServiceContext, userId);

                return wallet.TotalBalance >= transactionType.Fee;
            }
        }

        /// <summary>
        /// Gets the wallet for user.
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <returns>The wallet for the user</returns>
        public Wallet GetWalletForUser(Guid userId)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                return GetWalletForUserId(dbServiceContext, userId);
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
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                int count = dbServiceContext.Wallets.Count();

                if (count > 0)
                {
                    return 0;
                }

                int recordCount = 0;

                foreach (DataRow row in dataTable.Rows)
                {
                    int id = Convert.ToInt32(row["UserID"].ToString());

                    User user = dbServiceContext.User
                        .Include(u => u.Wallet)
                        .FirstOrDefault(u => u.ImportId == id);

                    if (user != null)
                    {
                        Wallet wallet = new Wallet
                        {
                            ImportId = Convert.ToInt32(row["ID"].ToString()),
                            TotalBalance = Convert.ToDouble(row["TotalBalance"].ToString()),
                            WalletTransactions = new List<WalletTransaction>()
                        };

                        user.Wallet = wallet;

                        recordCount++;
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
        /// Gets the type of the trans action.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <returns>The transaction type</returns>
        /// <exception cref="System.InvalidOperationException">Will be thrown if not found</exception>
        private static TransactionType GetTransActionType(DbServiceContext dbServiceContext, TransactionTypeNames transactionTypeName)
        {
            TransactionType transactionType = dbServiceContext.TransactionTypes
                .FirstOrDefault(t => t.Name == transactionTypeName.ToString());

            if (transactionType == null)
            {
                throw new InvalidOperationException(string.Format(Resource.ErrorTransactionTypeNotFound,
                    transactionTypeName));
            }

            return transactionType;
        }

        /// <summary>
        /// Gets the wallet for user identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>The wallet for the user id</returns>
        /// <exception cref="System.InvalidOperationException">Will be thrown if not found</exception>
        private static Wallet GetWalletForUserId(DbServiceContext dbServiceContext, Guid userId)
        {
            Wallet wallet = dbServiceContext.User
                .Include(u => u.Wallet)
                .Include(u => u.Wallet.WalletTransactions)
                .FirstOrDefault(u => u.Id.ToString() == userId.ToString())?.Wallet;

            if (wallet == null)
            {
                throw new InvalidOperationException(string.Format(Resource.ErrorWalletForUserIdNotFound, userId));
            }

            return wallet;
        }
    }
}
