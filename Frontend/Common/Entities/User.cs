using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations.Schema;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the user class
    /// </summary>
    [Table("Users")]
    public class User
    {
        /// <summary>
        /// Gets or sets the identifier.
        /// </summary>
        /// <value>
        /// The identifier.
        /// </value>
        public Guid Id { get; set; }

        /// <summary>
        /// Gets or sets the unique external user identifier.
        /// </summary>
        /// <value>
        /// The unique external user identifier.
        /// </value>
        /// <remarks>Mapping to the external authentication system unique user id</remarks>
        public Guid UniqueExternalUserId { get; set; }

        /// <summary>
        /// Gets or sets the name of the user.
        /// </summary>
        /// <value>
        /// The name of the user.
        /// </value>
        /// <remarks>Will be automatically created by the system</remarks>
        public string UserName { get; set; }

        /// <summary>
        /// Gets or sets the staked suggestions.
        /// </summary>
        /// <value>
        /// The staked suggestions.
        /// </value>
        /// <remarks>Only one suggestion can be staked per Issue</remarks>
        public List<Suggestion> StakedSuggestions { get; set; }
    }
}
