using System;
using System.ComponentModel.DataAnnotations;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the suggestion submission
    /// </summary>
    public class SuggestionSubmission
    {
        /// <summary>
        /// Gets or sets the identifier.
        /// </summary>
        /// <value>
        /// The identifier.
        /// </value>
        public Guid? Id { get; set; }

        /// <summary>
        /// Gets or sets the user identifier.
        /// </summary>
        /// <value>
        /// The user identifier.
        /// </value>
        [Required]
        public Guid UserId { get; set; }

        /// <summary>
        /// Gets or sets the issue identifier.
        /// </summary>
        /// <value>
        /// The issue identifier.
        /// </value>
        [Required]
        public Guid IssueId { get; set; }

        /// <summary>
        /// Gets or sets the title.
        /// </summary>
        /// <value>
        /// The title.
        /// </value>
        [Required]
        public string Title { get; set; }

        /// <summary>
        /// Gets or sets the description.
        /// </summary>
        /// <value>
        /// The description.
        /// </value>
        [Required]
        public string Description { get; set; }

        /// <summary>
        /// Converts to suggestion.
        /// </summary>
        /// <returns>The suggestion</returns>
        public Suggestion ToSuggestion()
        {
            Suggestion suggestion = new Suggestion
            {
                IssueId = IssueId,
                Title = Title,
                Description = Description,
                CreatorUserId = UserId
            };

            if (Id != null)
            {
                suggestion.Id = (Guid)Id;
            }

            return suggestion;
        }
    }
}
