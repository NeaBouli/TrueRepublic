using System;
using System.Collections.Generic;
using System.IO;
using Common.Data;
using Common.Entities;
using Common.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace PnyxWebAssembly.Server.Controllers
{
    /// <summary>
    /// Implementation of the issue controller
    /// </summary>
    /// <seealso cref="Microsoft.AspNetCore.Mvc.ControllerBase" />
    [ApiController]
    [Route("Issues")]
    public class IssueController : ControllerBase
    {
        /// <summary>
        /// The logger
        /// </summary>
        private readonly ILogger<IssueController> _logger;

        /// <summary>
        /// The configuration
        /// </summary>
        private readonly IConfiguration _configuration;

        /// <summary>
        /// Initializes a new instance of the <see cref="IssueController"/> class.
        /// </summary>
        /// <param name="logger">The logger.</param>
        /// <param name="configuration">The configuration.</param>
        public IssueController(ILogger<IssueController> logger, IConfiguration configuration)
        {
            _logger = logger;
            _configuration = configuration;

            DatabaseInitializationService.DbConnectString = configuration["DBConnectString"];
        }

        /// <summary>
        /// Gets all.
        /// </summary>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="userName">Name of the user.</param>
        /// <returns>
        /// All valid issues
        /// </returns>
        [HttpGet]
        public IActionResult GetAll([FromQuery] PaginatedList paginatedList, [FromQuery] string userName)
        {
            _logger.LogDebug(string.IsNullOrEmpty(userName) ? $"Get all issues for user {userName}" : "Get all issues");

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            decimal topStakedIssuesPercent = Convert.ToDecimal(_configuration["TopStakedIssuesPercent"]);

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                string userId = null;

                if (!string.IsNullOrEmpty(userName))
                {
                    userId = userService.GetUserId(dbServiceContext, userName);
                }

                IssueService issueService = new IssueService(topStakedIssuesPercent);

                List<Issue> issues = issueService.GetAll(dbServiceContext, paginatedList, userId);

                if (issues.Count == 0)
                {
                    return NotFound();
                }

                return Ok(issues);
            }
        }

        /// <summary>
        /// Gets the image for hashtags.
        /// </summary>
        /// <param name="hashtag">The hashtag.</param>
        /// <returns>
        /// The matching image for the given hashtags
        /// </returns>
        [HttpGet("ImageNameForHashtag/{hashtag}")]
        public IActionResult GetImageNameForHashtag(string hashtag)
        {
            _logger.LogDebug($"Get image name for {hashtag}");

            using DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();
            
            ImageInfoService imageInfoService = new ImageInfoService();

            string image = imageInfoService.GetImageForHashtag(dbServiceContext, hashtag);

            return Ok(image);
        }

        /// <summary>
        /// Gets the image.
        /// </summary>
        /// <param name="imageName">Name of the image.</param>
        /// <returns>The image for the given image name</returns>
        [HttpGet("Image/{imageName}")]
        public IActionResult GetImage(string imageName)
        {
            _logger.LogDebug($"Get image {imageName}");

            if (!System.IO.File.Exists(@$"Images\Cards\{imageName}"))
            {
                return NotFound();
            }

            return Ok(System.IO.File.Open(@$"Images\Cards\{imageName}", FileMode.Open, FileAccess.Read, FileShare.Read));
        }
    }
}
