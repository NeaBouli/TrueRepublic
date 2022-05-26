using System.ComponentModel.DataAnnotations;

namespace Pnyx.ApiClient.Interfaces.Transaction
{
    public interface ITransactionType
    {
        [Key]
        public Guid Id { get; set; }

        [Required]
        public string Name { get; set; }

        [Required]
        public double Fee { get; set; }
    }
}
