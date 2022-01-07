using System;
using System.Collections.Concurrent;
using System.IO;
using System.Net;
using System.Net.Http;
using System.Text.Encodings.Web;
using System.Threading.Tasks;

namespace PnyxWebAssembly.Client.Services
{
    /// <summary>
    /// Implementation of the issue image cache service
    /// </summary>
    public static class IssueImageCacheService
    {
        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        public static IHttpClientFactory ClientFactory { get; set; }

        /// <summary>
        /// The hashtags file name dictionary
        /// </summary>
        private static readonly ConcurrentDictionary<string, string> HashtagsFileNameDictionary = new();

        /// <summary>
        /// The file dictionary
        /// </summary>
        private static readonly ConcurrentDictionary<string, string> FileDictionary = new();

        /// <summary>
        /// Gets the image for hashtags.
        /// </summary>
        /// <param name="hashtags">The hashtags.</param>
        /// <returns>The image for the hashtags as base 64 decoded string</returns>
        public static async Task<string> GetImageForHashtags(string hashtags)
        {
            hashtags = hashtags.Replace(Environment.NewLine, string.Empty);
            hashtags = hashtags.Replace("\n", string.Empty);

            string imageName;

            using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");

            if (HashtagsFileNameDictionary.ContainsKey(hashtags))
            {
                imageName = HashtagsFileNameDictionary[hashtags];
            }
            else
            {
                imageName = await GetImageNameForHashtagsFromService(client, hashtags);

                HashtagsFileNameDictionary.TryAdd(hashtags, imageName);
            }

            string base64;

            if (FileDictionary.ContainsKey(imageName))
            {
                base64 = FileDictionary[imageName];
            }
            else
            {
                base64 = await GetImageFromService(client, imageName);
                FileDictionary.TryAdd(imageName, base64);
            }

            return base64;
        }

        /// <summary>
        /// Gets the image from service.
        /// </summary>
        /// <param name="client">The client.</param>
        /// <param name="imageName">Name of the image.</param>
        /// <returns>The image from the service</returns>
        private static async Task<string> GetImageFromService(HttpClient client, string imageName)
        {
            string base64;
            string contentType = Path.GetExtension(imageName).Replace(".", string.Empty);

            Stream imageStream = null;

            try
            {
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

            if (imageStream != null)
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
        /// Gets the image name for hashtags from service.
        /// </summary>
        /// <param name="client">The client.</param>
        /// <param name="hashtags">The hashtags.</param>
        /// <returns>The image name for hte given hashtags</returns>
        private static async Task<string> GetImageNameForHashtagsFromService(HttpClient client, string hashtags)
        {
            string hashtagsEncoded = UrlEncoder.Default.Encode(hashtags);

            string imageName;
            try
            {
                imageName = await client.GetStringAsync($"Issues/ImageNameForHashtags/{hashtagsEncoded}");
            }
            catch (HttpRequestException ex)
            {
                if (ex.StatusCode == HttpStatusCode.NotFound)
                {
                    imageName = "verträge.jpg";
                }
                else
                {
                    throw;
                }
            }

            return imageName;
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
