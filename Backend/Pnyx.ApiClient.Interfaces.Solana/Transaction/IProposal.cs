using System.ComponentModel.DataAnnotations;

namespace Pnyx.ApiClient.Interfaces.Transaction
{
    public interface IProposal
    {
        [Key]
        public Guid Id { get; set; }

        [Required]
        public Guid IssueId { get; set; }

        [Required]
        public string Title { get; set; }

        [Required]
        public string Description { get; set; }

        [Required]
        public DateTime CreateDate { get; set; }

        [Required]
        public Guid CreatorUserId { get; set; }

        public int StakeCount { get; set; }

        public bool IsStaked { get; }

        public bool IsTopStaked { get; }

        public bool HasMyStake { get; set; }

        public int? MyVote { get; set; }

        public bool HasMyVote { get; }

        public int VoteCount { get; set; }

        public bool IsVoted { get; }

        public bool IsTopVoted { get; set; }

        public bool CanEdit(Guid userId);
    }
}
