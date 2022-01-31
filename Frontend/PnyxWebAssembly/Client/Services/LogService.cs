using System.Net.Http;
using System.Threading.Tasks;
using Common.Entities;

namespace PnyxWebAssembly.Client.Services
{
    public static class LogService
    {
        /// <summary>
        /// Logs to server.
        /// </summary>
        /// <param name="client">The client.</param>
        /// <param name="logMessage">The log message.</param>
        public static async Task LogToServer(HttpClient client, string logMessage)
        {
            LogInfoItem logItem = new LogInfoItem(logMessage);
            MultipartFormDataContent content = new MultipartFormDataContent();
            content.Add(new StringContent(logItem.LogLevel.ToString()), "LogLevel");
            content.Add(new StringContent(logItem.LogText), "LogText");

            await client.PostAsync("Log", content);
        }
    }
}
