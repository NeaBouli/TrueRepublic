using System;
using System.ComponentModel.DataAnnotations;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the transaction type
    /// </summary>
    public class TransactionType
    {
        /// <summary>
        /// Gets or sets the identifier.
        /// </summary>
        /// <value>
        /// The identifier.
        /// </value>
        [Key]
        public Guid Id { get; set; }

        /// <summary>
        /// Gets or sets the name.
        /// </summary>
        /// <value>
        /// The name.
        /// </value>
        /// <remarks>Must be unique</remarks>
        [Required]
        public string Name { get; set; }

        /// <summary>
        /// Gets or sets the create issue.
        /// </summary>
        /// <value>
        /// The create issue.
        /// </value>
        [Required]
        public double Fee { get; set; }
    }
}
