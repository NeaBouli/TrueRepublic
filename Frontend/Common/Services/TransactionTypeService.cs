using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Data;
using Common.Entities;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the TransactionTypeDbService
    /// </summary>
    public class TransactionTypeService
    {
        public List<TransactionType> GetTransactionTypes(DbServiceContext dbServiceContext)
        {
            return dbServiceContext.TransactionTypes.ToList();
        }

        /// <summary>
        /// Gets the type of the transaction.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="transactionTypeName">Name of the transaction type.</param>
        /// <returns>The transaction type for the given transaction type name</returns>
        public TransactionType GetTransactionType(DbServiceContext dbServiceContext, TransactionTypeNames transactionTypeName)
        {
            return dbServiceContext.TransactionTypes
                .FirstOrDefault(t => t.Name == transactionTypeName.ToString());
        }

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>
        /// The number of imported records
        /// </returns>
        /// <exception cref="System.InvalidOperationException">Will be thrown if unknown transaction type name is used</exception>
        public int Import(DataTable dataTable)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                List<TransactionType> transactionTypes = GetTransactionTypes(dbServiceContext);

                if (transactionTypes.Count > 0)
                {
                    return 0;
                }

                int recordCount = 0;

                foreach (DataRow row in dataTable.Rows)
                {
                    TransactionType transactionType = new()
                    {
                        Name = row["Name"].ToString(),
                        ImportId = Convert.ToInt32(row["ID"].ToString()),
                        Fee = Convert.ToDouble(row["Fee"].ToString())
                    };

                    AddTransactionType(dbServiceContext, transactionType);

                    recordCount++;
                }

                if (recordCount > 0)
                {
                    dbServiceContext.SaveChanges();
                }

                return recordCount;
            }
        }

        /// <summary>
        /// Adds the specified transaction type.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="transactionType">Type of the transaction.</param>
        /// <exception cref="System.InvalidOperationException"></exception>
        public void AddTransactionType(DbServiceContext dbServiceContext, TransactionType transactionType)
        {
            if (!Enum.TryParse(typeof(TransactionTypeNames), transactionType.Name, true, out _))
            {
                throw new InvalidOperationException(
                    string.Format(Resource.ErrorUnknownTransactionTypeName, transactionType.Name));
            }

            dbServiceContext.TransactionTypes.Add(transactionType);
        }
    }
}
