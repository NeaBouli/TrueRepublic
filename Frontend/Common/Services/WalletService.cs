using System;
using System.Data;
using System.Linq;
using Common.Data;
using Common.Entities;
using Microsoft.EntityFrameworkCore;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the wallet service
    /// </summary>
    public class WalletService
    {
        /// <summary>
        /// Gets the total balance ot all wallets
        /// </summary>
        /// <returns>The total balance of all wallets</returns>
        public double GetTotalBalance(DbServiceContext dbServiceContext)
        {
            double totalBalance = 0;

            return Enumerable.Aggregate(dbServiceContext.Wallets, totalBalance,
                (current, wallet) => current + wallet.TotalBalance);
        }

        /// <summary>
        /// Gets the total balance.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>The total balance for the user -1 if not found</returns>
        public double GetTotalBalance(DbServiceContext dbServiceContext, string userId)
        {
            Wallet wallet = dbServiceContext.User
                .Include(u => u.Wallet).FirstOrDefault(u => u.Id.ToString() == userId)?.Wallet;

            if (wallet == null)
            {
                return -1;
            }

            return wallet.TotalBalance;
        }

        /// <summary>
        /// Determines whether [has enough funding for transaction] [the specified database service context].
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="userId">The user identifier.</param>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <returns>
        ///   <c>true</c> if [has enough funding for transaction] [the specified database service context]; otherwise, <c>false</c>.
        /// </returns>
        /// <exception cref="System.InvalidOperationException">Wallet for user {userId} not found</exception>
        private bool HasEnoughFundingForTransaction(DbServiceContext dbServiceContext, Guid userId,
            TransactionTypeNames transactionTypeName)
        {
            TransactionType transactionType = GetTransactionType(dbServiceContext, transactionTypeName);

            Wallet wallet = GetWalletForUserId(dbServiceContext, userId);

            if (wallet == null)
            {
                throw new InvalidOperationException($"Wallet for user {userId} not found");
            }

            return wallet.TotalBalance + transactionType.Fee >= 0;
        }

        /// <summary>
        /// Adds the transaction.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="userId">The user identifier.</param>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <param name="transactionId">The transaction identifier.</param>
        /// <exception cref="System.InvalidOperationException">
        /// Wallet for user {userId} not found
        /// </exception>
        public void AddTransaction(DbServiceContext dbServiceContext, Guid userId, TransactionTypeNames transactionTypeName, Guid? transactionId)
        {
            if (!HasEnoughFundingForTransaction(dbServiceContext, userId, transactionTypeName))
            {
                throw new InvalidOperationException(Resource.ErrorNotEnoughFounding);
            }

            Wallet wallet = GetWalletForUserId(dbServiceContext, userId);

            if (wallet == null)
            {
                throw new InvalidOperationException($"Wallet for user {userId} not found");
            }

            TransactionTypeService transactionTypeService = new TransactionTypeService();
            TransactionType transactionType = transactionTypeService.GetTransactionType(dbServiceContext, transactionTypeName);

            WalletTransaction walletTransaction = new WalletTransaction
            {
                Balance = transactionType.Fee,
                TransactionId = transactionId,
                TransactionType = transactionType,
                WalletId = wallet.Id
            };

            wallet.TotalBalance += transactionType.Fee;

            dbServiceContext.WalletTransactions.Add(walletTransaction);
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

            using DbServiceContext context = dbServiceContext;
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

                if (user != null && user.Wallet == null)
                {
                    Wallet wallet = new Wallet
                    {
                        ImportId = Convert.ToInt32(row["ID"].ToString()),
                        TotalBalance = Convert.ToDouble(row["TotalBalance"].ToString())
                    };

                    dbServiceContext.Wallets.Add(wallet);
                    user.WalletId = wallet.Id;

                    recordCount++;
                }
            }

            if (recordCount > 0)
            {
                dbServiceContext.SaveChanges();
            }

            return recordCount;
        }

        /// <summary>
        /// Gets the type of the trans action.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <returns>The transaction type</returns>
        /// <exception cref="System.InvalidOperationException">Will be thrown if not found</exception>
        private static TransactionType GetTransactionType(DbServiceContext dbServiceContext, TransactionTypeNames transactionTypeName)
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
