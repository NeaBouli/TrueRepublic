using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
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
        public static readonly Dictionary<string, string> HashtagsFileNameDictionary = new();

        /// <summary>
        /// The file dictionary
        /// </summary>
        public static readonly Dictionary<string, string> FileDictionary = new();

        /// <summary>
        /// Gets the image for hashtags.
        /// </summary>
        /// <param name="hashtags">The hashtags.</param>
        /// <returns>The image for the hashtags as base 64 decoded string</returns>
        public static async Task<string> GetImageForHashtags(string hashtags)
        {
            hashtags = hashtags.Replace(Environment.NewLine, string.Empty);
            hashtags = hashtags.Replace("\n", string.Empty);
            hashtags = hashtags.ToLowerInvariant();

            string image = null;

            using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");

            string[] hashtagsListRaw = hashtags.Split(new []{' ', '#'}, StringSplitOptions.RemoveEmptyEntries);

            List<string> hashtagsList = new List<string>();

            foreach (string hashtag in hashtagsListRaw)
            {
                hashtagsList.Add(!hashtag.StartsWith("#") ? $"#{hashtag}" : hashtag);
            }

            Dictionary<string, int> imageNamesCountDictionary = new Dictionary<string, int>();

            foreach (string hashtag in hashtagsList)
            {
                lock (HashtagsFileNameDictionary)
                {
                    if (HashtagsFileNameDictionary.ContainsKey(hashtag))
                    {
                        Debug.WriteLine($"Get hashtag {hashtag} from HashtagsFileNameDictionary");

                        string imageName = HashtagsFileNameDictionary[hashtag];

                        if (!imageNamesCountDictionary.ContainsKey(imageName))
                        {
                            imageNamesCountDictionary.Add(imageName, 0);
                        }
                        else
                        {
                            imageNamesCountDictionary[imageName]++;
                        }
                    }
                }
            }

            if (imageNamesCountDictionary.Count > 0)
            {
                image = imageNamesCountDictionary
                    .FirstOrDefault(i => i.Value == imageNamesCountDictionary.Values.Max()).Key;
            }

            if (string.IsNullOrEmpty(image))
            {
                foreach (string hashtag in hashtagsList)
                {
                    string imageName = await GetImageNameForHashtagFromService(client, hashtag);

                    if (!imageNamesCountDictionary.ContainsKey(imageName))
                    {
                        imageNamesCountDictionary.Add(imageName, 0);
                        lock (HashtagsFileNameDictionary)
                        {
                            Debug.WriteLine($"TryAdd {hashtag} {imageName}");
                            HashtagsFileNameDictionary.TryAdd(hashtag, imageName);
                            Debug.WriteLine($"HashtagsFileNameDictionary.Count: {HashtagsFileNameDictionary.Count}");
                        }
                    }
                    else
                    {
                        imageNamesCountDictionary[imageName]++;
                    }
                }
            }

            if (imageNamesCountDictionary.Count > 0)
            {
                image = imageNamesCountDictionary
                    .FirstOrDefault(i => i.Value == imageNamesCountDictionary.Values.Max()).Key;
            }

            if (string.IsNullOrEmpty(image))
            {
                image = "verträge.jpg";
            }

            string base64 = null;

            lock (FileDictionary)
            {
                if (FileDictionary.ContainsKey(image))
                {
                    base64 = FileDictionary[image];
                }
            }

            if (string.IsNullOrEmpty(base64))
            {
                base64 = await GetImageFromService(client, image);
                lock (FileDictionary)
                {
                    FileDictionary.TryAdd(image, base64);
                }
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
        /// <param name="hashtag">The hashtag.</param>
        /// <returns>
        /// The image name for hte given hashtags
        /// </returns>
        private static async Task<string> GetImageNameForHashtagFromService(HttpClient client, string hashtag)
        {
            string hashtagEncoded = UrlEncoder.Default.Encode(hashtag);

            string imageName;
            try
            {
                imageName = await client.GetStringAsync($"Issues/ImageNameForHashtag/{hashtagEncoded}");
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
