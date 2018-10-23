using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;

namespace cfusers.Controllers
{
    [Route("api/[controller]")]
    public class ErrorController : ControllerBase
    {
        [HttpGet]
        public async Task<IActionResult> GetUnauthorized()
        {
            return BadRequest("no.");
        }
    }
}