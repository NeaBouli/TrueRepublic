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
        private readonly ILogger<UserController> _logger;

        /// <summary>
        /// Initializes a new instance of the <see cref="IssueController"/> class.
        /// </summary>
        /// <param name="logger">The logger.</param>
        /// <param name="configuration">The configuration.</param>
        public UserController(ILogger<UserController> logger, IConfiguration configuration)
        {
            _logger = logger;

            DatabaseInitializationService.DbConnectString = configuration["DBConnectString"];
        }

        /// <summary>
        /// Gets the user by external identifier.
        /// </summary>
        /// <param name="externalUserId">The external user identifier.</param>
        /// <returns>The user if found otherwise null</returns>
        [HttpGet("ByExternalId/{externalUserId}")]
        public IActionResult GetUserByExternalId(string externalUserId)
        {
            _logger.LogDebug($"Get user by externalUserId {externalUserId}");

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                User user = userService.GetUserByExternalId(dbServiceContext, Guid.Parse(externalUserId));

                if (user == null)
                {
                    return NotFound();
                }

                return Ok(user);
            }
        }

        /// <summary>
        /// Gets the user by user identifier.
        /// </summary>
        /// <param name="userName">Name of the user.</param>
        /// <returns></returns>
        [HttpGet("ByName/{userName}")]
        public IActionResult GetUserByUserId(string userName)
        {
            _logger.LogDebug($"Get user by name {userName}");

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                User user = userService.GetUserByName(dbServiceContext, userName);

                if (user == null)
                {
                    return NotFound();
                }

                return Ok(user);
            }
        }

        /// <summary>
        /// Creates the specified user.
        /// </summary>
        /// <param name="user">The user.</param>
        /// <returns>Ok if user creation was successful</returns>
        [HttpPost]
        public IActionResult Create([FromBody] User user)
        {
            _logger.LogDebug($"Creating user {user.UserName}");

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                UserService userService = new UserService();

                userService.CreateUser(dbServiceContext, user, 100D);

                return Ok();
            }
        }
    }
}
