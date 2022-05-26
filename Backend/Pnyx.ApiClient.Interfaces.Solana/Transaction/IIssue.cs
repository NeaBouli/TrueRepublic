using System.ComponentModel.DataAnnotations;
using System.Text;

namespace Pnyx.ApiClient.Interfaces.Transaction
{
    public interface IIssue
    {
        [Key]
        public Guid Id { get; set; }

        [Required]
        public string Tags { get; set; }

        public string ImageName { get; set; }

        [Required]
        public string Title { get; set; }

        [Required]
        public string Description { get; set; }

        public DateTime? DueDate { get; set; }

        [Required]
        public DateTime CreateDate { get; set; }

        [Required]
        public Guid CreatorUserId { get; set; }

        public List<IProposal> Proposals { get; set; }

        public bool IsTopStaked { get; set; }

        public bool CanEdit(Guid userId);

        public bool HasMyStake { get; set; }

        public int TotalStakeCount { get; set; }

        public int TotalVoteCount { get; set; }

        public IEnumerable<string> GetTags();

        public bool HasTag(string tag);

        public void AddTag(string tag);

        public void RemoveTag(string tag);
    }
}
