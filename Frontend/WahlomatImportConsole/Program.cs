using System;
using System.Collections.Generic;
using System.Data;
using System.IO;
using System.Linq;
using System.Text;
using Common.Data;

namespace WahlomatImportConsole
{
    public class Program
    {
        public static void Main()
        {
            System.Text.Encoding.RegisterProvider(System.Text.CodePagesEncodingProvider.Instance);

            // TODO: get last index for issues  from excel
            ExcelDataContext.FullPathToXlsFile = "TestData.xlsx";
            ExcelDataContext excelDataContext = ExcelDataContext.GetInstance();

            DataTable dataTable = excelDataContext.Sheets["Issues"];

            if (dataTable == null)
            {
                Console.WriteLine(@"Could not find the issues sheet");
                return;
            }

            int maxId = 0;

            foreach (DataRow row in dataTable.Rows)
            {
                maxId = int.Parse(row["ID"].ToString() ?? "0");
            }

            if (maxId == 0)
            {
                Console.WriteLine(@"maxId could not be determined");
                return;
            }

            // TODO: read all module definitions into list with index, description & short description
            List<string> definitionFiles = new List<string> { "Files\\1\\module_definition.js", "Files\\2\\module_definition.js"};

            foreach (string definitionFile in definitionFiles)
            {
                int itemCount = 0;

                string shortDescription = string.Empty;

                foreach (string line in File.ReadLines(definitionFile, Encoding.UTF8))
                {
                    if (line.Contains($"WOMT_aThesen[{itemCount}][0][0]"))
                    {
                        shortDescription = line.Substring(line.IndexOf("='", StringComparison.Ordinal) + 1)
                            .Replace("';", string.Empty)
                            .Replace("'", string.Empty);
                    }
                    if (line.Contains($"WOMT_aThesen[{itemCount}][0][1]"))
                    {
                        var fullDescription = line.Substring(line.IndexOf("='", StringComparison.Ordinal) + 1)
                            .Replace("';", string.Empty)
                            .Replace("'", string.Empty);

                        if (!string.IsNullOrEmpty(shortDescription) && !string.IsNullOrEmpty(fullDescription))
                        {
                            maxId++;
                            
                            DataRow row = dataTable.NewRow();
                            row["ID"] = maxId;
                            row["Title"] = shortDescription;
                            row["Description"] = fullDescription;
                            row["CreatorUserID"] = 5;
                            dataTable.Rows.Add(row);

                            shortDescription = string.Empty;
                        }

                        itemCount++;
                    }
                }
            }

            StringBuilder sb = new StringBuilder();

            IEnumerable<string> columnNames = dataTable.Columns.Cast<DataColumn>().
            Select(column => column.ColumnName);
            sb.AppendLine(string.Join("|", columnNames));

            foreach (DataRow row in dataTable.Rows)
            {
                IEnumerable<string> fields = row.ItemArray.Select(field => field.ToString());
                sb.AppendLine(string.Join("|", fields));
            }

            File.WriteAllText("Issues.txt", sb.ToString(), Encoding.UTF8);
        }
    }
}
