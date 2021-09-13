using System;
using Common.Data;
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



            // TODO: import test data from excels if db was created - ask user which one to import

            // maybe put this one directly into the Gui
            // TODO: simulation: get issues, create issue, get suggestions, create suggestion, stake, show wallet, show transaction types
        }
    }
}
