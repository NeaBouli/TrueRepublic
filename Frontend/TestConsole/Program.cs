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
        public static void Main()
        {
            IConfigurationRoot config = new ConfigurationBuilder()
            .SetBasePath(AppDomain.CurrentDomain.BaseDirectory)
                .AddJsonFile("appsettings.json").Build();

            string dbConnectString = config["DBConnectString"];

            DatabaseInitializationService.DbConnectString = dbConnectString;
            DatabaseInitializationService.CreateDbIfNotExisting();

            ExcelImporterService excelImporterService = new ExcelImporterService();
            excelImporterService.ImportExcelFile("TestData.xlsx");

            // maybe put this one directly into the Gui
            // TODO: simulation: get issues, create issue, get suggestions, create suggestion, stake, show wallet, show transaction types
        }
    }
}
