using System.Collections.Generic;
using System.Data;
using Common.Entities;

namespace Common.Interfaces
{
    public interface ITransactionTypeInMemoryService
    {
        /// <summary>
        /// Gets the transaction types.
        /// </summary>
        /// <returns></returns>
        List<TransactionType> GetTransactionTypes();

        /// <summary>
        /// Gets the type of the transaction.
        /// </summary>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <returns>The transaction type for the given transaction type name</returns>
        TransactionType GetTransactionType(TransactionTypeNames transactionTypeName);

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        public void Import(DataTable dataTable);
    }
}