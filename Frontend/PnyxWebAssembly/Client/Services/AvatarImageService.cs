using System;
using System.IO;
using System.Net;
using System.Net.Http;
using System.Threading.Tasks;
using Common.Services;

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
            try
            {
                using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");

                await using Stream imageStream = await client.GetStreamAsync($"User/Avatar/{userName}");
                string contentType = await client.GetStringAsync($"User/AvatarContentType/{userName}");

                if (!string.IsNullOrEmpty(contentType))
                {
                    byte[] byteArray = ImageInfoService.StreamToByteArray(imageStream);

                    imageStream.Close();

                    string base64 = $"data:image/{contentType};base64, {Convert.ToBase64String(byteArray)}";

                    return base64;
                }
            }
            catch (HttpRequestException ex)
            {
                if (ex.StatusCode == HttpStatusCode.NotFound)
                {
                    return string.Empty;
                }

                throw;
            }

            return string.Empty;
        }
    }
}
