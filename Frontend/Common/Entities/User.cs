using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
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
        /// Initializes a new instance of the <see cref="User"/> class.
        /// </summary>
        public User()
        {
            Id = Guid.NewGuid();
        }

        /// <summary>
        /// Gets or sets the identifier.
        /// </summary>
        /// <value>
        /// The identifier.
        /// </value>
        [Key]
        public Guid Id { get; set; }

        /// <summary>
        /// Gets or sets the import identifier.
        /// </summary>
        /// <value>
        /// The import identifier.
        /// </value>
        public int? ImportId { get; set; }

        /// <summary>
        /// Gets or sets the unique external user identifier.
        /// </summary>
        /// <value>
        /// The unique external user identifier.
        /// </value>
        /// <remarks>Mapping to the external authentication system unique user id</remarks>
        [Required]
        public Guid UniqueExternalUserId { get; set; }

        /// <summary>
        /// Gets or sets the name of the user.
        /// </summary>
        /// <value>
        /// The name of the user.
        /// </value>
        /// <remarks>Will be automatically created by the system</remarks>
        [Required]
        public string UserName { get; set; }

        /// <summary>
        /// Gets or sets the wallet.
        /// </summary>
        /// <value>
        /// The wallet.
        /// </value>
        public Wallet Wallet { get; set; }

        /// <summary>
        /// Gets or sets the wallet identifier.
        /// </summary>
        /// <value>
        /// The wallet identifier.
        /// </value>
        public Guid? WalletId { get; set; }

        /// <summary>
        /// Gets or sets the staked suggestions.
        /// </summary>
        /// <value>
        /// The staked suggestions.
        /// </value>
        /// <remarks>Only one suggestion can be staked per Issue</remarks>
        public List<StakedSuggestion> StakedSuggestions { get; set; }
    }
}
