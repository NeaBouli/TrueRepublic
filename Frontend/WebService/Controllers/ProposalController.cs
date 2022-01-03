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
    /// Implementation of the Proposal controller
    /// </summary>
    [ApiController]
    [Route("Proposals")]
    public class ProposalController : ControllerBase
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
        /// Initializes a new instance of the <see cref="ProposalController"/> class.
        /// </summary>
        /// <param name="logger">The logger.</param>
        /// <param name="configuration">The configuration.</param>
        public ProposalController(ILogger<IssueController> logger, IConfiguration configuration)
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

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                string userId = userService.GetUserId(dbServiceContext, userName);

                ProposalService proposalService = new ProposalService(
                    Convert.ToInt32(_configuration["TopStakedProposalsPercent"]));

                List<Proposal> proposals = proposalService.GetByIssueId(dbServiceContext, id, userId);

                if (proposals == null)
                {
                    return NotFound();
                }

                return Ok(proposals);
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
        [HttpGet("Proposal/{id}")]
        public IActionResult GetByProposalId([FromQuery] string userName, string id)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                string userId = userService.GetUserId(dbServiceContext, userName);

                ProposalService proposalService = new ProposalService(
                    Convert.ToInt32(_configuration["TopStakedProposalsPercent"]));

                Proposal proposal = proposalService.GetByProposalId(dbServiceContext, id, userId);

                if (proposal == null)
                {
                    return NotFound();
                }

                return Ok(proposal);
            }
        }

        /// <summary>
        /// Adds the specified Proposal submission.
        /// </summary>
        /// <param name="proposalSubmission">The Proposal submission.</param>
        /// <returns>The id of the created Proposal</returns>
        [HttpPost]
        public IActionResult Add([FromBody] ProposalSubmission proposalSubmission)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                ProposalService proposalService = new ProposalService();

                Guid result = proposalService.Add(dbServiceContext, proposalSubmission);

                return Ok(result);
            }
        }

        /// <summary>
        /// Updates the specified Proposal submission.
        /// </summary>
        /// <param name="proposalSubmission">The Proposal submission.</param>
        /// <returns>The http status for the transaction</returns>
        [HttpPut]
        public IActionResult Update([FromBody] ProposalSubmission proposalSubmission)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                ProposalService proposalService = new ProposalService();

                proposalService.Update(dbServiceContext, proposalSubmission);

                return Ok();
            }
        }

        /// <summary>
        /// Stakes the Proposal.
        /// </summary>
        /// <param name="userIdItemId">The user identifier item identifier.</param>
        /// <returns>
        /// The http status for the transaction
        /// </returns>
        [HttpPut("StakeProposal")]
        public IActionResult StakeProposal([FromBody] UserIdItemId userIdItemId)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                StakedProposalService stakedProposalService =
                    new StakedProposalService(int.Parse(_configuration["ProposalStakeLifetimeDays"]));

                stakedProposalService.Stake(dbServiceContext, userIdItemId.ItemId, userIdItemId.UserId);

                return Ok();
            }
        }
    }
}
