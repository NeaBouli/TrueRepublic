using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Net.Http.Json;
using System.Threading.Tasks;
using Common.Entities;

namespace PnyxWebAssembly.Client.Services
{
    public class IssueService
    {
        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        public static IHttpClientFactory ClientFactory { get; set; }

        public void AddIssue(Issue issue)
        {
            throw new NoNullAllowedException();
        }

        public Issue GetIssue(string issueId)
        {
            throw new NoNullAllowedException();
        }

        /// <summary>
        /// Gets the hashtags.
        /// </summary>
        /// <param name="value">The value.</param>
        /// <returns>The hashtags for the given value if found</returns>
        public async Task<IEnumerable<string>> GetHashtags(string value)
        {
            if (string.IsNullOrEmpty(value))
            {
                return Array.Empty<string>();
            }

            if (value.StartsWith("#") && value.Length < 4 ||
                value.Length < 3)
            {
                return new List<string> { value };
            }

            using HttpClient client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Private");


            List<string> hashtags = await client.GetFromJsonAsync<List<string>>($"Issues/GetTagAutocomplete/{WebUtility.UrlEncode(value)}");

            if (hashtags == null || !hashtags.Any())
            {
                return new List<string> { value };
            }

            return hashtags;
        }
    }
}
