using System;
using System.ComponentModel.DataAnnotations;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the Proposal submission
    /// </summary>
    public class ProposalSubmission
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
        /// Converts to Proposal.
        /// </summary>
        /// <returns>The Proposal</returns>
        public Proposal ToProposal()
        {
            Proposal proposal = new Proposal
            {
                IssueId = IssueId,
                Title = Title,
                Description = Description,
                CreatorUserId = UserId
            };

            if (Id != null)
            {
                proposal.Id = (Guid)Id;
            }

            return proposal;
        }
    }
}
