using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the wallet
    /// </summary>
    public class Wallet
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="Wallet"/> class.
        /// </summary>
        public Wallet()
        {
            Id = Guid.NewGuid();
            WalletTransactions = new List<WalletTransaction>();
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
        /// Gets or sets the total balance.
        /// </summary>
        /// <value>
        /// The total balance.
        /// </value>
        [Required]
        public double TotalBalance { get; set; }

        /// <summary>
        /// Determines whether [has enough funding] [the specified balance].
        /// </summary>
        /// <param name="balance">The balance.</param>
        /// <returns>
        ///   <c>true</c> if [has enough funding] [the specified balance]; otherwise, <c>false</c>.
        /// </returns>
        public bool HasEnoughFunding(double balance)
        {
            return TotalBalance >= Math.Abs(balance);
        }

        /// <summary>
        /// Adds the transaction.
        /// </summary>
        /// <param name="walletTransaction">The wallet transaction.</param>
        /// <exception cref="System.InvalidOperationException">Not enough funding</exception>
        public void AddTransaction(WalletTransaction walletTransaction)
        {
            if (walletTransaction.Balance < 0 && !HasEnoughFunding(walletTransaction.Balance))
            {
                throw new InvalidOperationException(Resource.ErrorNotEnoughFounding);
            }

            WalletTransactions.Add(walletTransaction);

            TotalBalance += walletTransaction.Balance;
        }

        /// <summary>
        /// Gets or sets the wallet transactions.
        /// </summary>
        /// <value>
        /// The wallet transactions.
        /// </value>
        public List<WalletTransaction> WalletTransactions { get; set; }
    }
}
