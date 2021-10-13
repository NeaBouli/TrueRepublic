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
    /// Implementation of the suggestion controller
    /// </summary>
    [ApiController]
    [Route("Suggestions")]
    public class SuggestionController : ControllerBase
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
        /// Initializes a new instance of the <see cref="SuggestionController"/> class.
        /// </summary>
        /// <param name="logger">The logger.</param>
        /// <param name="configuration">The configuration.</param>
        public SuggestionController(ILogger<IssueController> logger, IConfiguration configuration)
        {
            _logger = logger;
            _configuration = configuration;

            DatabaseInitializationService.DbConnectString = configuration["DBConnectString"];
        }

        /// <summary>
        /// Gets the by identifier.
        /// </summary>
        /// <param name="userName">Name of the user.</param>
        /// <param name="id">The identifier.</param>
        /// <returns>
        /// The issue if found
        /// </returns>
        [HttpGet("Issue/{id}")]
        public IActionResult GetByIssueId([FromQuery] string userName, string id)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                string userId = UserService.GetUserId(dbServiceContext, userName);

                SuggestionService suggestionService = new SuggestionService(
                    Convert.ToInt32(_configuration["TopStakedSuggestionsPercent"]),
                    Convert.ToInt32(_configuration["TopVotedSuggestionsPercent"]));

                List<Suggestion> suggestions = suggestionService.GetByIssueId(dbServiceContext, id, userId);

                if (suggestions == null)
                {
                    return NotFound();
                }

                return Ok(suggestions);
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
        [HttpGet("Suggestion/{id}")]
        public IActionResult GetBySuggestionId([FromQuery] string userName, string id)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                string userId = UserService.GetUserId(dbServiceContext, userName);

                SuggestionService suggestionService = new SuggestionService(
                    Convert.ToInt32(_configuration["TopStakedSuggestionsPercent"]),
                    Convert.ToInt32(_configuration["TopVotedSuggestionsPercent"]));

                Suggestion suggestion = suggestionService.GetBySuggestionId(dbServiceContext, id, userId);

                if (suggestion == null)
                {
                    return NotFound();
                }

                return Ok(suggestion);
            }
        }

        /// <summary>
        /// Adds the specified suggestion submission.
        /// </summary>
        /// <param name="suggestionSubmission">The suggestion submission.</param>
        /// <returns>The id of the created suggestion</returns>
        [HttpPost]
        public IActionResult Add([FromBody] SuggestionSubmission suggestionSubmission)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                SuggestionService suggestionService = new SuggestionService();

                Guid result = suggestionService.Add(dbServiceContext, suggestionSubmission);

                return Ok(result);
            }
        }

        /// <summary>
        /// Updates the specified suggestion submission.
        /// </summary>
        /// <param name="suggestionSubmission">The suggestion submission.</param>
        /// <returns>The http status for the transaction</returns>
        [HttpPut]
        public IActionResult Update([FromBody] SuggestionSubmission suggestionSubmission)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                SuggestionService suggestionService = new SuggestionService();

                suggestionService.Update(dbServiceContext, suggestionSubmission);

                return Ok();
            }
        }

        /// <summary>
        /// Stakes the suggestion.
        /// </summary>
        /// <param name="userIdItemId">The user identifier item identifier.</param>
        /// <returns>
        /// The http status for the transaction
        /// </returns>
        [HttpPut("StakeSuggestion")]
        public IActionResult StakeSuggestion([FromBody] UserIdItemId userIdItemId)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                StakedSuggestionService stakedSuggestionService =
                    new StakedSuggestionService(int.Parse(_configuration["SuggestionStakeLifetimeDays"]));

                stakedSuggestionService.Stake(dbServiceContext, userIdItemId.ItemId, userIdItemId.UserId);

                return Ok();
            }
        }
    }
}
