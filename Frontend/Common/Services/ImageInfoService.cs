using System;
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
        public string GetImageForHashtags(DbServiceContext dbServiceContext, string hashtags)
        {
            throw new NotImplementedException();
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
