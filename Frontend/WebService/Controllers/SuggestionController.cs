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

        public SuggestionController(ILogger<IssueController> logger, IConfiguration configuration)
        {
            _logger = logger;
            _configuration = configuration;

            DatabaseInitializationService.DbConnectString = configuration["DBConnectString"];
        }

        /// <summary>
        /// Gets the by identifier.
        /// </summary>
        /// <param name="id">The identifier.</param>
        /// <returns>The issue if found</returns>
        [HttpGet("{id}")]
        public IActionResult GetById(string id)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                SuggestionService suggestionService = new SuggestionService();

                List<Suggestion> suggestions = suggestionService.GetById(dbServiceContext, id);

                if (suggestions == null)
                {
                    return Ok(new List<Suggestion>());
                }

                return Ok(suggestions);
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
    }
}
