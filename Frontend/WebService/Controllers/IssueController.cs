using System;
using System.Collections.Generic;
using Common.Data;
using Common.Entities;
using Common.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace WebService.Controllers
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

            string dockerEnvironmentConnectString = Environment.GetEnvironmentVariable("DBCONNECTSTRING_PNYX");

            if (!string.IsNullOrEmpty(dockerEnvironmentConnectString))
            {
                DatabaseInitializationService.DbConnectString = dockerEnvironmentConnectString;
                _logger.LogInformation($"Reading DB Connect string from Docker: {dockerEnvironmentConnectString}");
            }
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

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                string userId = userService.GetUserId(dbServiceContext, userName);

                IssueService issueService = new IssueService(topStakedIssuesPercent);

                return Ok(issueService.GetAll(dbServiceContext, paginatedList, userId));
            }
        }

        /// <summary>
        /// Gets the top staked.
        /// </summary>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="userName">Name of the user.</param>
        /// <returns>
        /// All valid issues
        /// </returns>
        [Route("GetTopStaked")]
        [HttpGet]
        public IActionResult GetTopStaked([FromQuery] PaginatedList paginatedList, [FromQuery] string userName)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            decimal topStakedIssuesPercent = Convert.ToDecimal(_configuration["TopStakedIssuesPercent"]);

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                string userId = userService.GetUserId(dbServiceContext, userName);

                IssueService issueService = new IssueService(topStakedIssuesPercent);

                return Ok(issueService.GetTopStaked(dbServiceContext, paginatedList, userId));
            }
        }

        /// <summary>
        /// Gets the issues by tags.
        /// </summary>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="userName">Name of the user.</param>
        /// <param name="tags">The tags.</param>
        /// <returns>
        /// All valid issues
        /// </returns>
        [Route("GetByTags/{tags}")]
        [HttpGet]
        public IActionResult GetByTags([FromQuery] PaginatedList paginatedList, [FromQuery] string userName, string tags)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                string userId = userService.GetUserId(dbServiceContext, userName);

                IssueService issueService = new IssueService();

                return Ok(issueService.GetByTags(dbServiceContext, tags, paginatedList, userId));
            }
        }

        /// <summary>
        /// Gets the top staked issues by tags.
        /// </summary>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="tags">The tags.</param>
        /// <param name="userName">Name of the user.</param>
        /// <returns>
        /// All valid issues
        /// </returns>
        [Route("GetTopStakedByTags/{tags}")]
        [HttpGet]
        public IActionResult GetTopStakedByTags([FromQuery] PaginatedList paginatedList, [FromQuery] string userName, string tags)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            decimal topStakedIssuesPercent = Convert.ToDecimal(_configuration["TopStakedIssuesPercent"]);

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                string userId = userService.GetUserId(dbServiceContext, userName);

                IssueService issueService = new IssueService(topStakedIssuesPercent);

                return Ok(issueService.GetTopStakedByTags(dbServiceContext, tags, paginatedList, userId));
            }
        }

        /// <summary>
        /// Gets the by identifier.
        /// </summary>
        /// <param name="userName">Name of the user.</param>
        /// <param name="id">The identifier.</param>
        /// <returns>
        /// The issue if found
        /// </returns>
        [HttpGet("{id}")]
        public IActionResult GetById([FromQuery] string userName, string id)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                string userId = userService.GetUserId(dbServiceContext, userName);

                IssueService issueService = new IssueService();

                Issue issue = issueService.GetById(dbServiceContext, id, userId);

                if (issue == null)
                {
                    return NotFound();
                }

                return Ok(issue);
            }
        }

        /// <summary>
        /// Adds the specified user issue.
        /// </summary>
        /// <param name="issueSubmission">The user issue.</param>
        /// <returns>The id of the created issue</returns>
        [HttpPost]
        public IActionResult Add([FromBody] IssueSubmission issueSubmission)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                IssueService issueService = new IssueService();

                Guid result = issueService.Add(dbServiceContext, issueSubmission);

                return Ok(result);
            }
        }

        /// <summary>
        /// Updates the specified issue.
        /// </summary>
        /// <param name="issueSubmission">The issue.</param>
        /// <returns>
        /// http status ok if everything went well
        /// </returns>
        [HttpPut]
        public IActionResult Update([FromBody] IssueSubmission issueSubmission)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                IssueService issueService = new IssueService();

                issueService.Update(dbServiceContext, issueSubmission);

                return Ok();
            }
        }

        [Route("GetTagAutocomplete/{tag}")]
        [HttpGet]
        public IActionResult GetTagAutocomplete(string tag)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                IssueService issueService = new IssueService();

                List<string> tags = issueService.GetTagAutocomplete(dbServiceContext, tag);

                return Ok(tags);
            }
        }
    }
}
