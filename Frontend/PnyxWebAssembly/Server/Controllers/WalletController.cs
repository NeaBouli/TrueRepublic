using Common.Data;
using Common.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace PnyxWebAssembly.Server.Controllers
{
    [ApiController]
    [Route("Wallets")]
    public class WalletController : ControllerBase
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
        public WalletController(ILogger<IssueController> logger, IConfiguration configuration)
        {
            _logger = logger;
            _configuration = configuration;

            DatabaseInitializationService.DbConnectString = configuration["DBConnectString"];
        }

        /// <summary>
        /// Gets the total balance.
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// The total balance for the given user id
        /// </returns>
        [HttpGet("GetTotalBalance/{userId}")]
        public IActionResult GetTotalBalance(string userId)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                WalletService walletService = new WalletService();

                double totalBalance = walletService.GetTotalBalance(dbServiceContext, userId);

                if ((int)totalBalance == -1)
                {
                    return NotFound();
                }

                return Ok(totalBalance);
            }
        }
    }
}
