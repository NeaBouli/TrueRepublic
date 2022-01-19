using System;
using System.IO;
using System.Net;
using System.Net.Http;
using System.Threading.Tasks;

namespace PnyxWebAssembly.Client.Services
{
    /// <summary>
    /// Implementation of the avatar image cache service
    /// </summary>
    public static class AvatarImageService
    {
        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        public static IHttpClientFactory ClientFactory { get; set; }

        /// <summary>
        /// Gets the avatar image base64.
        /// </summary>
        /// <param name="userName">Name of the user.</param>
        /// <returns></returns>
        public static async Task<string> GetAvatarImageBase64(string userName)
        {
            // TODO: implement cache as described here: https://docs.microsoft.com/de-de/aspnet/core/blazor/state-management?view=aspnetcore-6.0&pivots=webassembly#browser-storage-wasm

            using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");

            Stream imageStream = null;

            try
            {
                imageStream = await client.GetStreamAsync($"User/Avatar/{userName}");
            }
            catch (HttpRequestException ex)
            {
                if (ex.StatusCode == HttpStatusCode.NotFound)
                {
                    imageStream?.Close();

                    imageStream = null;
                }
                else
                {
                    throw;
                }
            }

            string contentType;

            try
            {
                contentType = await client.GetStringAsync($"User/AvatarContentType/{userName}");
            }
            catch (HttpRequestException ex)
            {
                if (ex.StatusCode == HttpStatusCode.NotFound)
                {
                    contentType = string.Empty;
                }
                else
                {
                    throw;
                }
            }

            if (imageStream != null && !string.IsNullOrEmpty(contentType))
            {
                byte[] byteArray = StreamToByteArray(imageStream);

                imageStream.Close();

                string base64 = $"data:image/{contentType};base64, {Convert.ToBase64String(byteArray)}";

                return base64;
            }

            return string.Empty;
        }

        /// <summary>
        /// Converts a stream to a byte array
        /// </summary>
        /// <param name="input">The input.</param>
        /// <returns>A byte array for the given stream</returns>
        private static byte[] StreamToByteArray(Stream input)
        {
            byte[] buffer = new byte[16 * 1024];
            using MemoryStream ms = new MemoryStream();
            int read;
            while ((read = input.Read(buffer, 0, buffer.Length)) > 0)
            {
                ms.Write(buffer, 0, read);
            }
            return ms.ToArray();
        }
    }
}
