using System;
using System.IO;
using System.Net;
using System.Net.Http;
using System.Threading.Tasks;
using Common.Services;

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
        /// Gets or sets the image cache service.
        /// </summary>
        /// <value>
        /// The image cache service.
        /// </value>
        public static  ImageCacheService ImageCacheService { get; set; }

        /// <summary>
        /// Gets the image from service.
        /// </summary>
        /// <param name="issueId">Name of the image.</param>
        /// <returns>
        /// The image from the service
        /// </returns>
        public static async Task<string> GetImageFromService(Guid issueId)
        {
            string contentType = null;

            Stream imageStream = null;

            try
            {
                using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");

                string imageName = await client.GetStringAsync($"Issues/ImageNameForIssue/{issueId}");

                if (!ImageCacheService.HasImage(imageName))
                {
                    contentType = Path.GetExtension(imageName).Replace(".", string.Empty);
                    imageStream = await client.GetStreamAsync($"Issues/Image/{imageName}");
                }
                else
                {
                    string imageData = ImageCacheService.Get(imageName);

                    if (!string.IsNullOrEmpty(imageData))
                    {
                        await LogService.LogToServer(client, $"Getting image {imageName} from cache");

                        return imageData;
                    }
                }

                if (!string.IsNullOrEmpty(contentType))
                {
                    byte[] byteArray = ImageInfoService.StreamToByteArray(imageStream);

                    imageStream.Close();

                    string base64 = $"data:image/{contentType};base64, {Convert.ToBase64String(byteArray)}";

                    if (!string.IsNullOrEmpty(imageName) && !ImageCacheService.HasImage(imageName))
                    {
                        ImageCacheService.Add(imageName, base64);
                    }

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
