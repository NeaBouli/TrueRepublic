using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Linq;

namespace Common.Entities
{
    // TODO: what status can an issue have: staking, voting, closed
    // TODO: who changes the status at which condition

    /// <summary>
    /// Implementation of the issue class
    /// </summary>
    /// <remarks>Record cannot be changed after it was created. Creator will not be tracked</remarks>
    public class Issue
    {
        /// <summary>Initializes a new instance of the <see cref="Issue" /> class.</summary>
        public Issue()
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
        // TODO: unique
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
        /// Gets or sets the create date.
        /// </summary>
        /// <value>
        /// The create date.
        /// </value>
        [Required]
        public DateTime CreateDate { get; set; }

        /// <summary>
        /// Gets or sets the creator user identifier.
        /// </summary>
        /// <value>
        /// The creator user identifier.
        /// </value>
        [Required]
        public Guid CreatorUserId { get; set; }

        /// <summary>
        /// Gets the Proposals.
        /// </summary>
        /// <value>
        /// The Proposals.
        /// </value>
        public List<Proposal> Proposals { get; set; }

        /// <summary>
        /// Gets or sets a value indicating whether this instance is top staked.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance is top staked; otherwise, <c>false</c>.
        /// </value>
        [NotMapped]
        public bool IsTopStaked { get; set; }

        /// <summary>
        /// Determines whether this instance can edit the specified user identifier.
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        ///   <c>true</c> if this instance can edit the specified user identifier; otherwise, <c>false</c>.
        /// </returns>
        public bool CanEdit(Guid userId)
        {
            if (userId.ToString() == CreatorUserId.ToString() &&
                (Proposals == null || Proposals.Count == 0))
            {
                return true;
            }

            return false;
        }

        /// <summary>
        /// Determines whether [has my stake].
        /// </summary>
        /// <returns>
        ///   <c>true</c> if [has my stake]; otherwise, <c>false</c>.
        /// </returns>
        public bool HasMyStake()
        {
            return Proposals.Any(proposal => proposal.HasMyStake);
        }

        /// <summary>
        /// Gets the total stake count.
        /// </summary>
        /// <returns>The total stake count for all assigned stakes</returns>
        public int GetTotalStakeCount()
        {
            return Proposals.Sum(proposal => proposal.StakeCount);
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
    }
}
