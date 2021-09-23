using System;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace Common.Entities
{
    public class StakedSuggestion
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="StakedSuggestion"/> class.
        /// </summary>
        public StakedSuggestion()
        {
            Id = Guid.NewGuid();
            CreateDate = DateTime.Now;
        }

        /// <summary>
        /// Gets or sets the create date.
        /// </summary>
        /// <value>
        /// The create date.
        /// </value>
        [DatabaseGenerated(DatabaseGeneratedOption.Computed)]
        public DateTime CreateDate { get; set; }

        /// <summary>
        /// Gets the valid till.
        /// </summary>
        /// <value>
        /// The valid till.
        /// </value>
        public DateTime ValidTill => CreateDate.AddDays(30);

        /// <summary>
        /// Gets a value indicating whether this instance is expired.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance is expired; otherwise, <c>false</c>.
        /// </value>
        [NotMapped]
        public bool IsExpired => ValidTill < DateTime.Now;

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
        /// Gets or sets the issue identifier.
        /// </summary>
        /// <value>
        /// The issue identifier.
        /// </value>
        [Required]
        public Guid IssueId { get; set; }

        /// <summary>
        /// Gets or sets the suggestion.
        /// </summary>
        /// <value>
        /// The suggestion.
        /// </value>
        public Suggestion Suggestion { get; set; }
    }
}
