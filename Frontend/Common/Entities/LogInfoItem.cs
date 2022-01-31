using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;

namespace Common.Entities
{
    public class LogInfoItem
    {
        public LogInfoItem()
        {
            LogLevel = LogLevel.Information;
        }

        public LogInfoItem(string logText)
        {
            LogLevel = LogLevel.Information;
            LogText = logText;
        }

        public LogInfoItem(LogLevel logLevel, string logText)
        {
            LogLevel = logLevel;
            LogText = logText;
        }

        public LogLevel LogLevel { get; set; }

        public string LogText { get; set; }
    }
}
