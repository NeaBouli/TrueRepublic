using System;
using System.Collections.Generic;
using System.Data;
using Common.Entities;

namespace Common.Interfaces
{
    public interface IWalletTransactionService
    {
        /// <summary>
        /// Gets the wallet transactions for user.
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <param name="fromDate">From date.</param>
        /// <param name="toDate">To date.</param>
        /// <returns>The wallet transactions for the user</returns>
        List<WalletTransaction> GetWalletTransactionsForUser(Guid userId, DateTime? fromDate = null, DateTime? toDate = null);

        /// <summary>
        /// Adds the wallet transaction.
        /// </summary>
        /// <param name="wallet">The wallet.</param>
        /// <param name="walletTransaction">The wallet transaction.</param>
        void AddWalletTransaction(Wallet wallet, WalletTransaction walletTransaction);

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>The number of imported records</returns>
        int Import(DataTable dataTable);
    }
}