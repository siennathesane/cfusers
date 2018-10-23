using System;
using System.Collections.Generic;
using System.Globalization;
using System.Linq;
using System.Threading.Tasks;

namespace cfusers.Models
{
    public class User
    {
        public string GivenName;
        public string FamilyName;
        public string Email;
        public DateTime DateStart;
        public string KeepAlive;
        public string DefaultPassword;
        private bool userExists;
        private bool cloudFoundryOrgExists;
        private bool cloudFoundrySpaceExists;
        private string shortenedName;

        private string uaaUserId;
        private string cfOrgId;
        private string cfSpaceId;

        public string CfSpaceId { get => cfSpaceId; set => cfSpaceId = value; }
        public string CfOrgId { get => cfOrgId; set => cfOrgId = value; }
        public string UaaUserId { get => uaaUserId; set => uaaUserId = value; }
        public bool CloudFoundryOrgExists { get => cloudFoundryOrgExists; set => cloudFoundryOrgExists = value; }
        public bool CloudFoundrySpaceExists { get => cloudFoundrySpaceExists; set => cloudFoundrySpaceExists = value; }
        public string ShortenedName { get => shortenedName; set => shortenedName = value; }
        public bool UserExists { get => userExists; set => userExists = value; }

        public User(string firstName, string lastName, string email, DateTime dateStart, string keepAlive, string defaultPassword)
        {
            this.GivenName = firstName;
            this.FamilyName = lastName;
            this.Email = email;
            // make sure any password gets set.
            if (defaultPassword == "")
            {
                this.DefaultPassword = Environment.GetEnvironmentVariable("DEFAULT_PASSWORD");
            }
            else
            {
                this.DefaultPassword = defaultPassword;
            }
            this.UserExists = false;
            this.CloudFoundryOrgExists = false;
            this.CloudFoundrySpaceExists = false;

        }

        public DateTime ParseDate(string dateStart)
        {
            string[] formatRefs = { "yyyy-MM-ddTHH:mm:ss.fffZ" };
            try
            {
                DateTime parsedDate = DateTime.ParseExact(dateStart, formatRefs[0], CultureInfo.InvariantCulture, DateTimeStyles.None);
                return parsedDate;
            }
            catch (FormatException fe)
            {
                throw fe;
            }
        }

        public async Task CreateUserAsync(User newUser)
        {

        }
    }
}
