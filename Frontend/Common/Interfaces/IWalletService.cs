using System;
using System.Data;
using Common.Entities;

namespace Common.Interfaces
{
    public interface IWalletService
    {
        /// <summary>
        /// Gets the total balance ot all wallets
        /// </summary>
        /// <returns>The total balance of all wallets</returns>
        double GetTotalBalance();

        /// <summary>
        /// Determines whether [has enough funding for transaction] [the specified user identifier].
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <returns>
        ///   <c>true</c> if [has enough funding for transaction] [the specified user identifier]; otherwise, <c>false</c>.
        /// </returns>
        bool HasEnoughFundingForTransaction(Guid userId, TransactionTypeNames transactionTypeName);

        /// <summary>
        /// Gets the wallet for user.
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <returns>The wallet for the user</returns>
        Wallet GetWalletForUser(Guid userId);

        /// <summary>
        /// Adds the transaction.
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <param name="transactionId">The transaction identifier.</param>
        void AddTransaction(Guid userId, TransactionTypeNames transactionTypeName, Guid? transactionId);

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>
        /// The number of imported records
        /// </returns>
        int Import(DataTable dataTable);
    }
}