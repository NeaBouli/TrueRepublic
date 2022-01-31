using Common.Entities;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging;

namespace PnyxWebAssembly.Server.Controllers
{
    [ApiController]
    [Route("Log")]
    public class LogController : ControllerBase
    {
        /// <summary>
        /// The logger
        /// </summary>
        private readonly ILogger<LogController> _logger;

        /// <summary>
        /// Initializes a new instance of the <see cref="LogController"/> class.
        /// </summary>
        /// <param name="logger">The logger.</param>
        public LogController(ILogger<LogController> logger)
        {
            _logger = logger;
        }

        /// <summary>
        /// Logs the information.
        /// </summary>
        /// <param name="logInfoItem">The log information item.</param>
        /// <returns></returns>
        [HttpPost]

        public IActionResult Log([FromForm] LogInfoItem logInfoItem)
        {
            _logger.Log(logInfoItem.LogLevel, $"[CLIENT] {logInfoItem.LogText}");

            return Ok();
        }
    }
}
