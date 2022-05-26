using System.ComponentModel.DataAnnotations;
using Pnyx.ApiClient.Interfaces.Transaction;
using Pnyx.ApiClient.Interfaces.Wallet;

namespace Pnyx.ApiClient.Interfaces.Session
{
    public interface IUser
    {
        [Key]
        public Guid Id { get; set; }

        [Required]
        public Guid UniqueExternalUserId { get; set; }

        [Required]
        public string UserName { get; set; }

        public IWallet Wallet { get; set; }

        public List<IStakedProposal> StakedSuggestions { get; set; }

        public List<IVote> Votes { get; set; }
    }
}
