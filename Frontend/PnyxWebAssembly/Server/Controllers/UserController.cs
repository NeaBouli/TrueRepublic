using System;
using System.IO;
using Common.Data;
using Common.Entities;
using Common.Services;
using Microsoft.AspNetCore.Authorization;
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
        [Authorize]
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
        /// Gets the user by identifier.
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <returns>The user by id</returns>
        [Authorize]
        [HttpGet("ById/{userId}")]
        public IActionResult GetUserById(string userId)
        {
            _logger.LogDebug($"Get user by id {userId}");

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            UserService userService = new UserService();

            using (dbServiceContext)
            {
                User user = userService.GetUserById(dbServiceContext, Guid.Parse(userId));

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
        /// <returns>The user by the given name</returns>
        [Authorize]
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
        [Authorize]
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

        /// <summary>
        /// Gets the avatar.
        /// </summary>
        /// <param name="userName">Name of the user.</param>
        /// <returns>The avatar image stream</returns>
        [HttpGet("Avatar/{userName}")]
        public IActionResult GetAvatar(string userName)
        {
            if (!System.IO.File.Exists(@$"Images\Avatars\{userName}.jpg") &&
                !System.IO.File.Exists(@$"Images\Avatars\{userName}.png"))
            {
                return NotFound();
            }

            string imageName = $"{userName}.jpg";

            if (!System.IO.File.Exists(@$"Images\Avatars\{imageName}"))
            {
                imageName = $"{userName}.png";
            }

            return Ok(System.IO.File.Open(@$"Images\Avatars\{imageName}", FileMode.Open, FileAccess.Read, FileShare.Read));
        }

        /// <summary>
        /// Gets the type of the avatar content.
        /// </summary>
        /// <param name="userName">Name of the user.</param>
        /// <returns>The avatar image type</returns>
        [HttpGet("AvatarContentType/{userName}")]
        public IActionResult GetAvatarContentType(string userName)
        {
            if (!System.IO.File.Exists(@$"Images\Avatars\{userName}.jpg") &&
                !System.IO.File.Exists(@$"Images\Avatars\{userName}.png"))
            {
                return NotFound();
            }

            string imageName = $"{userName}.jpg";
            string imageType = "jpeg";

            if (!System.IO.File.Exists(@$"Images\Avatars\{imageName}"))
            {
                imageType = "png";
            }

            return Ok(imageType);
        }
    }
}
