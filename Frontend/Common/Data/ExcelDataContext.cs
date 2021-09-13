using System.Data;
using System.IO;
using ExcelDataReader;

namespace Common.Data
{
    /// <summary>
    /// Implementation of the excel data context
    /// </summary>
    public class ExcelDataContext
    {
        /// <summary>
        /// The instance
        /// </summary>
        private static ExcelDataContext _instance;

        /// <summary>
        /// Prevents a default instance of the <see cref="ExcelDataContext"/> class from being created.
        /// </summary>
        private ExcelDataContext()
        {
            FileStream stream = File.Open(FullPathToXlsFile, FileMode.Open, FileAccess.Read);
            IExcelDataReader excelReader = ExcelReaderFactory.CreateOpenXmlReader(stream);

            DataSet result = excelReader.AsDataSet(new ExcelDataSetConfiguration()
            {
                ConfigureDataTable = (_) => new ExcelDataTableConfiguration()
                {
                    UseHeaderRow = true
                }
            });

            Sheets = result.Tables;
        }

        /// <summary>
        /// Gets the instance.
        /// </summary>
        /// <returns></returns>
        public static ExcelDataContext GetInstance() => _instance ??= new ExcelDataContext();

        /// <summary>
        /// Gets or sets the full path to XLS file.
        /// </summary>
        /// <value>
        /// The full path to XLS file.
        /// </value>
        public static string FullPathToXlsFile { get; set; }

        /// <summary>
        /// Gets the sheets.
        /// </summary>
        /// <value>
        /// The sheets.
        /// </value>
        public DataTableCollection Sheets { get; }
    }
}

