using System;
using System.Collections.Generic;
using System.Linq;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using cfusers.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace cfusers.Controllers
{
    [Route("api/[controller]")]
    public class UserController : Controller
    {
        private readonly ILogger<UserController> logger;
        private UserContext userContext;
        
        public UserController(ILoggerFactory fact, UserContext userContext)
        {
            this.logger = fact.CreateLogger<UserController>();
            this.userContext = userContext;
        }

        [HttpGet("{id}")]
        public async Task<IActionResult> GetUser(string email)
        {
            User localUserRef = await this.userContext.Users.FirstOrDefaultAsync(o => o.Email == email);
            if (localUserRef.Email == "")
            {
                return NotFound();
            }
            return Ok(localUserRef);
        }

        [HttpGet]
        public async Task<IActionResult> GetUsers()
        {
            List<User> localUserRefs = await this.userContext.Users.ToListAsync();
            int usersFound = localUserRefs.Capacity;
            this.logger.LogInformation(@"found {usersFound} users.");
            return Ok(localUserRefs);
        }

        [HttpPost]
        public async Task<IActionResult> CreateUser(User newUser)
        {
            if (newUser == null)
            {
                return BadRequest();
            }

        }
    }
}