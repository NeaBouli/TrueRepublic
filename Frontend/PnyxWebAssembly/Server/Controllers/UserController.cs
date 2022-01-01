using System;
using Common.Data;
using Common.Entities;
using Common.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace PnyxWebAssembly.Server.Controllers
{
    [ApiController]
    [Route("User")]
    public class UserController : ControllerBase
    {
        /// <summary>
        /// The logger
        /// </summary>
        private readonly ILogger<IssueController> _logger;


        /// <summary>
        /// Initializes a new instance of the <see cref="IssueController"/> class.
        /// </summary>
        /// <param name="logger">The logger.</param>
        /// <param name="configuration">The configuration.</param>
        public UserController(ILogger<IssueController> logger, IConfiguration configuration)
        {
            _logger = logger;

            DatabaseInitializationService.DbConnectString = configuration["DBConnectString"];
        }

        /// <summary>
        /// Gets the user by external identifier.
        /// </summary>
        /// <param name="externalUserId">The external user identifier.</param>
        /// <returns>The user if found otherwise null</returns>
        [HttpGet("{externalUserId}")]
        public IActionResult GetUserByExternalId(string externalUserId)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                User user = userService.GetUserByExternalId(dbServiceContext, Guid.Parse(externalUserId));

                return Ok(user);
            }
        }
    }
}
