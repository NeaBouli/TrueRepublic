using System;
using Common.Services;
using Microsoft.Extensions.Configuration;

namespace TestConsole
{
    /// <summary>
    /// Implementation of the program
    /// </summary>
    public static class Program
    {
        /// <summary>
        /// Defines the entry point of the application.
        /// </summary>
        public static void Main()
        {
            IConfigurationRoot config = new ConfigurationBuilder()
            .SetBasePath(AppDomain.CurrentDomain.BaseDirectory)
                .AddJsonFile("appsettings.json").Build();

            string dbConnectString = config["DBConnectString"];

            DatabaseInitializationService.DbConnectString = dbConnectString;
            DatabaseInitializationService.CreateDbIfNotExisting();

            ExcelImporterService excelImporterService = new ExcelImporterService();
            excelImporterService.TableImported += ExcelImporterService_TableImported;
            excelImporterService.ImportExcelFile("TestData.xlsx");
        }

        /// <summary>
        /// Handles the TableImported event of the ExcelImporterService control.
        /// </summary>
        /// <param name="sender">The source of the event.</param>
        /// <param name="e">The <see cref="Common.Entities.NameCountEventArgs"/> instance containing the event data.</param>
        private static void ExcelImporterService_TableImported(object sender, Common.Entities.NameCountEventArgs e)
        {
            Console.WriteLine(@$"{e.Name} successful imported - {e.Count} records");
        }
    }
}
