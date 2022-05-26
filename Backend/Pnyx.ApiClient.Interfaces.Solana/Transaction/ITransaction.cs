using System.ComponentModel.DataAnnotations;

namespace Pnyx.ApiClient.Interfaces.Transaction
{
    public interface ITransaction
    {
        [Key]
        public Guid Id { get; set; }

        [Required]
        public Guid WalletId { get; set; }

        [Required]
        public double Balance { get; set; }

        [Required]
        public ITransactionType TransactionType { get; set; }

        [Required]
        public Guid TransactionTypeId { get; set; }

        [Required]
        public DateTime CreateDate { get; set; }
    }
}
