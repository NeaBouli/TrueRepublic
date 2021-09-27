using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the user issue
    /// </summary>
    /// <seealso cref="Common.Entities.Issue" />
    public class IssueSubmission
    {
        /// <summary>
        /// Gets or sets the user identifier.
        /// </summary>
        /// <value>
        /// The user identifier.
        /// </value>
        [Required]
        public Guid UserId { get; set; }

        /// <summary>
        /// Gets or sets the tags.
        /// </summary>
        /// <value>
        /// The tags.
        /// </value>
        [Required]
        public string Tags { get; set; }

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

        // TODO: remove and put in interval (snapshot)

        /// <summary>
        /// Gets or sets the due date.
        /// </summary>
        /// <value>
        /// The due date.
        /// </value>
        public DateTime? DueDate { get; set; }

        /// <summary>
        /// Converts to issue.
        /// </summary>
        /// <returns></returns>
        public Issue ToIssue()
        {
            Issue issue = new Issue
            {
                Suggestions = new List<Suggestion>(),
                DueDate = DueDate,
                Description = Description,
                Tags = Tags,
                Title = Title
            };

            return issue;
        }
    }
}
