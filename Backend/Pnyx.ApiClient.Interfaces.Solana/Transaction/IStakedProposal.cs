using System.ComponentModel.DataAnnotations;

namespace Pnyx.ApiClient.Interfaces.Transaction
{
    public interface IStakedProposal
    {
        [Required]
        public DateTime CreateDate { get; set; }

        public DateTime ValidUntil { get; }

        public bool IsExpired { get; }

        [Key]
        public Guid Id { get; set; }

        [Required]
        public Guid IssueId { get; set; }

        [Required]
        public Guid UserId { get; set; }

        public IProposal Proposal { get; set; }

        [Required]
        public Guid ProposalId { get; set; }

        [Required]
        public int ExpirationDays { get; set; }
    }
}
