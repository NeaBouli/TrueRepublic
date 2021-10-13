using System;
using System.ComponentModel.DataAnnotations;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the vote
    /// </summary>
    public class Vote
    {
        /// <summary>
        /// The value
        /// </summary>
        private int _value;

        /// <summary>
        /// Initializes a new instance of the <see cref="Vote"/> class.
        /// </summary>
        public Vote()
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
        /// Gets or sets the import identifier.
        /// </summary>
        /// <value>
        /// The import identifier.
        /// </value>
        public int? ImportId { get; set; }

        /// <summary>
        /// Gets or sets the wallet identifier.
        /// </summary>
        /// <value>
        /// The wallet identifier.
        /// </value>
        [Required]
        public Guid UserId { get; set; }

        /// <summary>
        /// Gets or sets the suggestion identifier.
        /// </summary>
        /// <value>
        /// The suggestion identifier.
        /// </value>
        [Required]
        public Guid IssueId { get; set; }

        /// <summary>
        /// Gets or sets the suggestion identifier.
        /// </summary>
        /// <value>
        /// The suggestion identifier.
        /// </value>
        [Required]
        public Guid SuggestionId { get; set; }

        /// <summary>
        /// Gets or sets the suggestion.
        /// </summary>
        /// <value>
        /// The suggestion.
        /// </value>
        public Suggestion Suggestion { get; set; }

        /// <summary>
        /// Gets or sets the value.
        /// </summary>
        /// <value>
        /// The value.
        /// </value>
        [Required, Range(-5, 5)]
        public int Value
        {
            get => _value;
            set
            {
                _value = value;
                LastModifiedDate = DateTime.Now;
            }
        }

        /// <summary>
        /// Gets or sets the create date.
        /// </summary>
        /// <value>
        /// The create date.
        /// </value>
        [Required]
        public DateTime CreateDate { get; set; }

        /// <summary>
        /// Gets or sets the create date.
        /// </summary>
        /// <value>
        /// The create date.
        /// </value>
        [Required]
        public DateTime LastModifiedDate { get; set; }
    }
}
