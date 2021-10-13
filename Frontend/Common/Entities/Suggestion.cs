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
        /// The is top staked
        /// </summary>
        private bool _isTopStaked;

        /// <summary>
        /// The is top voted
        /// </summary>
        private bool _isTopVoted;

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

        /// <summary>
        /// Gets or sets a value indicating whether this instance is top staked.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance is top staked; otherwise, <c>false</c>.
        /// </value>
        [NotMapped]
        public bool IsTopStaked
        {
            get => IsStaked && _isTopStaked;

            set => _isTopStaked = value;
        }

        /// <summary>
        /// Gets or sets a value indicating whether this instance has my stake.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance has my stake; otherwise, <c>false</c>.
        /// </value>
        [NotMapped]
        public bool HasMyStake { get; set; }

        /// <summary>
        /// Gets or sets my vote.
        /// </summary>
        /// <value>
        /// My vote.
        /// </value>
        [NotMapped, Range(-5, 5)]
        public int? MyVote { get; set; }

        /// <summary>
        /// Gets a value indicating whether this instance has my vote.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance has my vote; otherwise, <c>false</c>.
        /// </value>
        [NotMapped]
        public bool HasMyVote => MyVote != null;

        /// <summary>
        /// Gets or sets the vote count.
        /// </summary>
        /// <value>
        /// The vote count.
        /// </value>
        [NotMapped]
        public int VoteCount { get; set; }

        /// <summary>
        /// Gets a value indicating whether this instance is voted.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance is voted; otherwise, <c>false</c>.
        /// </value>
        [NotMapped]
        public bool IsVoted => VoteCount > 0;

        /// <summary>
        /// Gets or sets a value indicating whether this instance is top voted.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance is top voted; otherwise, <c>false</c>.
        /// </value>
        [NotMapped]
        public bool IsTopVoted
        {
            get => IsVoted && _isTopVoted;

            set => _isTopVoted = value;
        }

        /// <summary>
        /// Determines whether this instance can edit the specified user identifier.
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        ///   <c>true</c> if this instance can edit the specified user identifier; otherwise, <c>false</c>.
        /// </returns>
        public bool CanEdit(Guid userId)
        {
            if (userId.ToString() == CreatorUserId.ToString() && !IsStaked)
            {
                return true;
            }

            return false;
        }
    }
}
