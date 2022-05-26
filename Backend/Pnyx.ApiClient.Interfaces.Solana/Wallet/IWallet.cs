using System.ComponentModel.DataAnnotations;

namespace Pnyx.ApiClient.Interfaces.Wallet
{
    public interface IWallet
    {
        [Key]
        public Guid Id { get; set; }

        [Required]
        public Guid WalletId { get; set; }

        [Required]
        public double Balance { get; set; }

        public Guid? TransactionId { get; set; }

        public DateTime CreateDate { get; set; }
    }
}
