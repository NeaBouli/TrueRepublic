using System;
using System.ComponentModel.DataAnnotations;

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
        public DateTime CreateDate { get; set; }

        /// <summary>
        /// Gets the valid till.
        /// </summary>
        /// <value>
        /// The valid till.
        /// </value>
        public DateTime ValidTill => CreateDate.AddDays(30);

        /// <summary>
        /// Gets or sets the identifier.
        /// </summary>
        /// <value>
        /// The identifier.
        /// </value>
        [Key]
        public Guid Id { get; set; }

        /// <summary>
        /// Gets or sets the suggestion.
        /// </summary>
        /// <value>
        /// The suggestion.
        /// </value>
        public Suggestion Suggestion { get; set; }
    }
}
