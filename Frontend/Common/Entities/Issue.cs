using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Linq;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the issue class
    /// </summary>
    /// <remarks>Record cannot be changed after it was created. Creator will not be tracked</remarks>
    [Table("Issues")]
    public class Issue
    {
        /// <summary>Initializes a new instance of the <see cref="Issue" /> class.</summary>
        public Issue()
        {
            Id = Guid.NewGuid();
            CreateDate = DateTime.Now;
        }

        /// <summary>
        /// Initializes a new instance of the <see cref="Issue"/> class.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <param name="title">The title.</param>
        /// <param name="description">The description.</param>
        /// <param name="dueDate">The due date.</param>
        public Issue(string tags, string title, string description, DateTime? dueDate = null)
        {
            Id = Guid.NewGuid();
            CreateDate = DateTime.Now;

            Tags = tags;
            Title = title;
            Description = description;
            DueDate = dueDate;
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

        /// <summary>
        /// Gets or sets the due date.
        /// </summary>
        /// <value>
        /// The due date.
        /// </value>
        public DateTime? DueDate { get; set; }

        /// <summary>
        /// Gets or sets the create date.
        /// </summary>
        /// <value>
        /// The create date.
        /// </value>
        [DatabaseGenerated(DatabaseGeneratedOption.Computed)]
        public DateTime CreateDate { get; set; }

        /// <summary>
        /// Gets the error message.
        /// </summary>
        /// <value>
        /// The error message.
        /// </value>
        [NotMapped]
        public string ErrorMessage { get; private set; }

        /// <summary>
        /// Gets the suggestions.
        /// </summary>
        /// <value>
        /// The suggestions.
        /// </value>
        public List<Suggestion> Suggestions => new List<Suggestion>();

        /// <summary>
        /// Gets the total stake count.
        /// </summary>
        /// <returns>The total stake count for all assigned stakes</returns>
        public int GetTotalStakeCount()
        {
            return Suggestions.Sum(suggestion => suggestion.StakeCount);
        }

        /// <summary>
        /// Gets the tags.
        /// </summary>
        /// <returns>The tags</returns>
        public IEnumerable<string> GetTags()
        {
            return GetTags(Tags);
        }

        /// <summary>
        /// Gets the tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <returns>The tags as list</returns>
        public static IEnumerable<string> GetTags(string tags)
        {
            string[] tagItems = tags.Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);

            foreach (string tag in tagItems)
            {
                yield return tag;
            }
        }

        /// <summary>
        /// Determines whether the specified tag has tag.
        /// </summary>
        /// <param name="tag">The tag.</param>
        /// <returns>
        ///   <c>true</c> if the specified tag has tag; otherwise, <c>false</c>.
        /// </returns>
        public bool HasTag(string tag)
        {
            List<string> tags = new List<string>(GetTags());

            return tags.Any(tagFromList => 
                string.Equals(tag, tagFromList, StringComparison.OrdinalIgnoreCase) || 
                string.Equals($"#{tag}", tagFromList, StringComparison.OrdinalIgnoreCase));
        }

        /// <summary>
        /// Gets the top suggestions.
        /// </summary>
        /// <param name="limit">The limit.</param>
        /// <returns>The top suggestions from the list</returns>
        public List<Suggestion> GetTopSuggestions(int limit)
        {
            if (limit >= Suggestions.Count)
            {
                return Suggestions;
            }

            return Suggestions.OrderByDescending(s => s.StakeCount).ThenBy(t => t.CreateDate).Take(limit).ToList();
        }

        /// <summary>
        /// Gets the top suggestions percentage.
        /// </summary>
        /// <param name="percentage">The percentage.</param>
        /// <param name="maxCount">The maximum count.</param>
        /// <returns>
        /// The top suggestions from the list
        /// </returns>
        /// <remarks>Example: green area has to be chosen that only 10% (x %) are in red limit to 12</remarks>
        public List<Suggestion> GetTopSuggestionsPercentage(decimal percentage, int maxCount = 0)
        {
            if (percentage >= 100)
            {
                return Suggestions;
            }

            decimal count = Convert.ToDecimal(Suggestions.Count);

            int limit = Convert.ToInt32(Math.Round(percentage / 100 * count));

            if (maxCount > 0 && maxCount < limit)
            {
                limit = maxCount;
            }

            return Suggestions.OrderByDescending(s => s.StakeCount).ThenBy(t => t.CreateDate).Take(limit).ToList();
        }

        /// <summary>
        /// Returns true if ... is valid.
        /// </summary>
        /// <returns>
        ///   <c>true</c> if this instance is valid; otherwise, <c>false</c>.
        /// </returns>
        public bool IsValid()
        {
            // TODO: put into service - entity should be data only

            if (string.IsNullOrEmpty(Tags))
            {
                ErrorMessage = "Tags are required";
                return false;
            }

            if (string.IsNullOrEmpty(Title))
            {
                ErrorMessage = "Title is required";
                return false;
            }

            if (string.IsNullOrEmpty(Description))
            {
                ErrorMessage = "Description is required";
                return false;
            }

            if (!string.IsNullOrEmpty(Tags) && Tags.Length < 5)
            {
                ErrorMessage = "Tags must be >= 5 characters";
                return false;
            }

            if (!string.IsNullOrEmpty(Title) && Title.Length < 5)
            {
                ErrorMessage = "Title must be >= 5 characters";
                return false;
            }

            if (!string.IsNullOrEmpty(Description) && Description.Length < 5)
            {
                ErrorMessage = "Description must be >= 5 characters";
                return false;
            }

            if (DueDate != null && DueDate < DateTime.Now)
            {
                ErrorMessage = "Due Date must be in future";
                return false;
            }

            if (DueDate != null)
            {
                double differenceDays = ((DateTime)DueDate).Subtract(DateTime.Now).TotalDays;

                if (differenceDays < 5)
                {
                    ErrorMessage = "Due Date must be at least 5 days in the future";
                    return false;
                }
            }

            ErrorMessage = string.Empty;
            return true;
        }
    }
}
