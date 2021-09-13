using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Common.Data;
using Common.Entities;
using Common.Interfaces;

namespace Common.Services
{
    public class TransactionTypeDbService : ITransactionTypeInMemoryService
    {
        public static string DbConnectString { get; set; }

        public TransactionTypeDbService()
        {
            if (string.IsNullOrEmpty(DbConnectString))
            {
                throw new InvalidOperationException(Resource.ErrorDbConnectStringCannotBeEmpty);
            }

            DbServiceContext dbServiceContext = new DbServiceContext(DbConnectString);

            using (dbServiceContext)
            {
                dbServiceContext.Database.EnsureCreated();
            }
        }
        public List<TransactionType> GetTransactionTypes()
        {
            throw new NotImplementedException();
        }

        public TransactionType GetTransactionType(TransactionTypeNames transactionTypeName)
        {
            throw new NotImplementedException();
        }

        public void Import(DataTable dataTable)
        {
            throw new NotImplementedException();
        }
    }
}
