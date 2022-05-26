using System.ComponentModel.DataAnnotations;

namespace Pnyx.ApiClient.Interfaces.Transaction
{
    public interface IVote
    {
        [Key]
        public Guid Id { get; set; }

        [Required]
        public Guid UserId { get; set; }

        [Required]
        public Guid IssueId { get; set; }

        [Required]
        public Guid ProposalId { get; set; }

        public IProposal Proposal { get; set; }

        [Required]
        public int Value { get; set; }

        [Required]
        public DateTime CreateDate { get; set; }

        [Required]
        public DateTime LastModifiedDate { get; set; }
    }
}
