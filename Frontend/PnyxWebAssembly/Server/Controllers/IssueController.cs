using System;
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
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            decimal topStakedIssuesPercent = Convert.ToDecimal(_configuration["TopStakedIssuesPercent"]);

            using (dbServiceContext)
            {
                string userId = UserService.GetUserId(dbServiceContext, userName);

                IssueService issueService = new IssueService(topStakedIssuesPercent);

                return Ok(issueService.GetAll(dbServiceContext, paginatedList, userId));
            }
        }
    }
}
