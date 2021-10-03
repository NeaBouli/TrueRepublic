using System;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the user id / item id
    /// </summary>
    public class UserIdItemId
    {
        /// <summary>
        /// Gets or sets the user identifier.
        /// </summary>
        /// <value>
        /// The user identifier.
        /// </value>
        public Guid UserId { get; set; }

        /// <summary>
        /// Gets or sets the item identifier.
        /// </summary>
        /// <value>
        /// The item identifier.
        /// </value>
        public Guid ItemId { get; set; }
    }
}
