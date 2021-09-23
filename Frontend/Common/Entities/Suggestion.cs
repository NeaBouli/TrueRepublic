using System;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the suggestion
    /// </summary>
    [Table("Suggestions")]
    public class Suggestion
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="Suggestion"/> class.
        /// </summary>
        public Suggestion()
        {
            Id = Guid.NewGuid();
            CreateDate = DateTime.Now;
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
        /// Gets or sets the issue identifier.
        /// </summary>
        /// <value>
        /// The issue identifier.
        /// </value>
        [Required]
        public Guid IssueId { get; set; }

        /// <summary>
        /// Gets or sets the import identifier.
        /// </summary>
        /// <value>
        /// The import identifier.
        /// </value>
        public int? ImportId { get; set; }

        /// <summary>
        /// Gets or sets the short description.
        /// </summary>
        /// <value>
        /// The short description.
        /// </value>
        public string Title { get; set; }

        /// <summary>
        /// Gets or sets the description.
        /// </summary>
        /// <value>
        /// The description.
        /// </value>
        public string Description { get; set; }

        /// <summary>
        /// Gets or sets the create date.
        /// </summary>
        /// <value>
        /// The create date.
        /// </value>
        [DatabaseGenerated(DatabaseGeneratedOption.Computed)]
        public DateTime CreateDate { get; set; }

        /// <summary>
        /// Gets or sets the stake count.
        /// </summary>
        /// <value>
        /// The stake count.
        /// </value>
        [NotMapped]
        public int StakeCount { get; set; }

        /// <summary>
        /// Gets a value indicating whether this instance is staked.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance is staked; otherwise, <c>false</c>.
        /// </value>
        [NotMapped]
        public bool IsStaked => StakeCount > 0;
    }
}
