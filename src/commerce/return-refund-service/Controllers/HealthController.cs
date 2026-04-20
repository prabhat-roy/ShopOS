using Microsoft.AspNetCore.Mvc;

namespace ReturnRefundService.Controllers;

[ApiController]
[Route("healthz")]
public class HealthController : ControllerBase
{
    /// <summary>Liveness probe endpoint.</summary>
    [HttpGet]
    [ProducesResponseType(StatusCodes.Status200OK)]
    public IActionResult GetHealth()
    {
        return Ok(new { status = "ok" });
    }
}
