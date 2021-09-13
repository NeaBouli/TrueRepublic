using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Entities;
using Common.Interfaces;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the TransactionTypeInMemoryService
    /// </summary>
    public class TransactionTypeInMemoryService : ITransactionTypeInMemoryService
    {
        /// <summary>
        /// The transaction types
        /// </summary>
        private readonly List<TransactionType> _transactionTypes;

        /// <summary>
        /// Initializes a new instance of the <see cref="TransactionTypeInMemoryService"/> class.
        /// </summary>
        /// <param name="transactionTypes">The transaction types.</param>
        public TransactionTypeInMemoryService(List<TransactionType> transactionTypes)
        {
            _transactionTypes = transactionTypes;
        }

        /// <summary>
        /// Gets the transaction types.
        /// </summary>
        /// <returns></returns>
        public List<TransactionType> GetTransactionTypes()
        {
            return _transactionTypes;
        }

        /// <summary>
        /// Gets the type of the transaction.
        /// </summary>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <returns>The transaction type for the given transaction type name</returns>
        public TransactionType GetTransactionType(TransactionTypeNames transactionTypeName)
        {
            return _transactionTypes.FirstOrDefault(t => t.Name == transactionTypeName.ToString());
        }

        public void Import(DataTable dataTable)
        {
            throw new System.NotImplementedException();
        }
    }
}
