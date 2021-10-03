using System.Collections.Generic;
using Common.Data;
using Common.Entities;
using Common.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace WebService.Controllers
{
    [ApiController]
    [Route("Wallets")]
    public class WalletController : ControllerBase
    {
        /// <summary>
        /// The logger
        /// </summary>
        private readonly ILogger<WalletController> _logger;

        /// <summary>
        /// The configuration
        /// </summary>
        private readonly IConfiguration _configuration;

        /// <summary>
        /// Initializes a new instance of the <see cref="IssueController"/> class.
        /// </summary>
        /// <param name="logger">The logger.</param>
        /// <param name="configuration">The configuration.</param>
        public WalletController(ILogger<WalletController> logger, IConfiguration configuration)
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

        /// <summary>
        /// Gets the wallet transactions.
        /// </summary>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="limit">The limit.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns></returns>
        [HttpGet("GetWalletTransactions/{userId}")]
        public IActionResult GetWalletTransactions([FromQuery] PaginatedList paginatedList, [FromQuery] int limit,
            string userId)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                WalletTransactionService walletTransactionService = new WalletTransactionService();

                List<WalletTransaction> walletTransactions = walletTransactionService.GetWalletTransactionsForUser(
                    dbServiceContext, userId, paginatedList, limit);

                if (walletTransactions == null)
                {
                    return NotFound();
                }

                return Ok(walletTransactions);
            }
        }
    }
}
