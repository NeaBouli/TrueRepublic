using System;
using System.Data;
using Common.Data;
using Common.Entities;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the ExcelImporterService
    /// </summary>
    public class ExcelImporterService
    {
        /// <summary>
        /// Occurs when [table imported].
        /// </summary>
        public event EventHandler<NameCountEventArgs> TableImported;

        /// <summary>
        /// Imports the excel file.
        /// </summary>
        /// <param name="fullPathToExcelFile">The full path to excel file.</param>
        /// <exception cref="System.NotImplementedException">Will be thrown if unknown sheet is there</exception>
        public void ImportExcelFile(string fullPathToExcelFile)
        {
            // TODO: walk pages
            ExcelDataContext.FullPathToXlsFile = fullPathToExcelFile;
            ExcelDataContext excelDataContext = ExcelDataContext.GetInstance();

            // TODO: check if db table is empty
            foreach (DataTable sheet in excelDataContext.Sheets)
            {
                switch (sheet.TableName)
                {
                    case "TransactionTypes":
                        ImportTransactionTypes(sheet);
                        break;
                    case "Users":
                        ImportUsers(sheet);
                        break;
                    case "Wallets":
                        ImportWallets(sheet);
                        break;
                    case "WalletTransactions":
                        ImportWalletTransactions(sheet);
                        break;
                    case "Issues":
                        ImportIssues(sheet);
                        break;
                    case "Suggestions":
                        ImportSuggestions(sheet);
                        break;
                    case "StakedSuggestions":
                        ImportStakedSuggestions(sheet);
                        break;
                    default:
                        throw new NotImplementedException();
                }
            }
        }

        private void ImportStakedSuggestions(DataTable sheet)
        {
            throw new NotImplementedException();
        }

        private void ImportSuggestions(DataTable sheet)
        {
            throw new NotImplementedException();
        }

        private void ImportIssues(DataTable sheet)
        {
            throw new NotImplementedException();
        }

        private void ImportWalletTransactions(DataTable sheet)
        {
            throw new NotImplementedException();
        }

        private void ImportWallets(DataTable sheet)
        {
            throw new NotImplementedException();
        }

        /// <summary>
        /// Imports the users.
        /// </summary>
        /// <param name="sheet">The sheet.</param>
        private void ImportUsers(DataTable sheet)
        {
            UserService userService = new UserService();
            OnTableImported(new NameCountEventArgs("Users", userService.Import(sheet)));
        }

        /// <summary>
        /// Imports the transaction types.
        /// </summary>
        /// <param name="sheet">The sheet.</param>
        private void ImportTransactionTypes(DataTable sheet)
        {
            TransactionTypeService transactionTypeDbService = new();
            OnTableImported(new NameCountEventArgs("ImportTransactionTypes", transactionTypeDbService.Import(sheet)));
        }

        /// <summary>
        /// Raises the <see cref="E:TableImported" /> event.
        /// </summary>
        /// <param name="e">The <see cref="NameCountEventArgs"/> instance containing the event data.</param>
        protected virtual void OnTableImported(NameCountEventArgs e)
        {
            TableImported?.Invoke(this, e);

        }
    }
}
