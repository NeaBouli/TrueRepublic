using System.Collections.Concurrent;

namespace PnyxWebAssembly.Client.Services
{
    /// <summary>
    /// Implementation of the image cache service
    /// </summary>
    public class ImageCacheService
    {
        /// <summary>
        /// The image cache
        /// </summary>
        private readonly ConcurrentDictionary<string, string> _imageCache = new();

        /// <summary>
        /// Adds the specified name.
        /// </summary>
        /// <param name="name">The name.</param>
        /// <param name="data">The data.</param>
        public void Add(string name, string data)
        {
            _imageCache.TryAdd(name, data);
        }

        /// <summary>
        /// Determines whether the specified name has image.
        /// </summary>
        /// <param name="name">The name.</param>
        /// <returns>
        ///   <c>true</c> if the specified name has image; otherwise, <c>false</c>.
        /// </returns>
        public bool HasImage(string name)
        {
            return _imageCache.ContainsKey(name);
        }

        /// <summary>
        /// Gets the specified name.
        /// </summary>
        /// <param name="name">The name.</param>
        /// <returns></returns>
        public string Get(string name)
        {
            _imageCache.TryGetValue(name, out string value);

            return value;
        }
    }
}
