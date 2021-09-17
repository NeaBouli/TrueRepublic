using System;
using System.ComponentModel.DataAnnotations;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the wallet transaction
    /// </summary>
    public class WalletTransaction
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WalletTransaction"/> class.
        /// </summary>
        public WalletTransaction()
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
        /// Gets or sets the import identifier.
        /// </summary>
        /// <value>
        /// The import identifier.
        /// </value>
        public int? ImportId { get; set; }

        /// <summary>
        /// Gets or sets the balance.
        /// </summary>
        /// <value>
        /// The balance.
        /// </value>
        [Required]
        public double Balance { get; set; }

        /// <summary>
        /// Gets or sets the reference action.
        /// </summary>
        /// <value>
        /// The reference action.
        /// </value>
        /// <remarks>Might be an enum instead. Idea is to reference the action that triggered the operation</remarks>
        [Required] 
        public TransactionType TransactionType { get; set; }

        /// <summary>
        /// Gets or sets the transaction identifier.
        /// </summary>
        /// <value>
        /// The transaction identifier.
        /// </value>
        /// <remarks>
        /// Optional. May not set every time due to privacy.
        /// Issues will not be set but stakes will
        /// </remarks>
        public Guid? TransactionId { get; set; }

        /// <summary>
        /// Gets or sets the create date.
        /// </summary>
        /// <value>
        /// The create date.
        /// </value>
        [Required]
        public DateTime CreateDate { get; set; }
    }
}
