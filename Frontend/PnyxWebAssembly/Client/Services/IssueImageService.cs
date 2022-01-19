using System;
using System.IO;
using System.Net;
using System.Net.Http;
using System.Threading.Tasks;

namespace PnyxWebAssembly.Client.Services
{
    /// <summary>
    /// Implementation of the issue image cache service
    /// </summary>
    public static class IssueImageService
    {
        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        public static IHttpClientFactory ClientFactory { get; set; }

        /// <summary>
        /// Gets the image from service.
        /// </summary>
        /// <param name="issueId">Name of the image.</param>
        /// <returns>
        /// The image from the service
        /// </returns>
        public static async Task<string> GetImageFromService(Guid issueId)
        {
            string base64;
            string contentType = null;
            
            Stream imageStream = null;

            // TODO: implement cache as described here: https://docs.microsoft.com/de-de/aspnet/core/blazor/state-management?view=aspnetcore-6.0&pivots=webassembly#browser-storage-wasm

            using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");

            try
            {
                string imageName = await client.GetStringAsync($"Issues/ImageNameForIssue/{issueId}");
                contentType = Path.GetExtension(imageName).Replace(".", string.Empty);
                imageStream = await client.GetStreamAsync($"Issues/Image/{imageName}");
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

            if (!string.IsNullOrEmpty(contentType) && imageStream != null)
            {
                byte[] byteArray = StreamToByteArray(imageStream);

                imageStream.Close();

                base64 = $"data:image/{contentType};base64, {Convert.ToBase64String(byteArray)}";
            }
            else
            {
                base64 = string.Empty;
            }

            return base64;
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
