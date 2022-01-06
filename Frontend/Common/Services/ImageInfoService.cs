using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Data;
using Common.Entities;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the image info service
    /// </summary>
    public class ImageInfoService
    {
        /// <summary>
        /// Gets the image for hashtags.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="hashtags">The hashtags.</param>
        /// <returns>The image for the hashtags</returns>
        public string GetImageForHashtags(DbServiceContext dbServiceContext, string hashtags)
        {
            IEnumerable<string> tags = Issue.GetTags(hashtags);

            List<string> hashTagsList = new List<string>();

            foreach (string tag in tags)
            {
                hashTagsList.Add(tag.StartsWith("#") ? tag : $"#{tag}");
            }

            Dictionary<string, int> imageNamesCountDictionary = new Dictionary<string, int>();

            foreach (string hashTag in hashTagsList)
            {
                List<ImageInfo> images = 
                    dbServiceContext.ImageInfos.Where(i => i.Hashtags.Contains(hashTag)).ToList();

                foreach (ImageInfo image in images)
                {
                    if (!imageNamesCountDictionary.ContainsKey(image.Filename))
                    {
                        imageNamesCountDictionary.Add(image.Filename, 0);
                    }
                    else
                    {
                        imageNamesCountDictionary[image.Filename]++;
                    }
                }
            }

            string fileName = imageNamesCountDictionary
                .FirstOrDefault(i => i.Value == imageNamesCountDictionary.Values.Max()).Key;

            if (string.IsNullOrEmpty(fileName))
            {
                fileName = "verträge.jpg";
            }

            return fileName;
        }

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>
        /// The number of imported records
        /// </returns>
        public int Import(DataTable dataTable)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using DbServiceContext context = dbServiceContext;
            int count = dbServiceContext.ImageInfos.Count();

            if (count > 0)
            {
                return 0;
            }

            int recordCount = 0;

            foreach (DataRow row in dataTable.Rows)
            {
                ImageInfo imageInfo = new ImageInfo
                {
                    Hashtags = row["Hashtags"].ToString(),
                    Filename = row["Filename"].ToString()
                };

                dbServiceContext.ImageInfos.Add(imageInfo);

                recordCount++;
            }

            if (recordCount > 0)
            {
                dbServiceContext.SaveChanges();
            }

            return recordCount;
        }
    }
}
