using Common.Entities;
using Microsoft.EntityFrameworkCore;

namespace Common.Data
{
    public class DbServiceContext : DbContext
    {
        /// <summary>
        /// The connect string
        /// </summary>
        public static string ConnectString { get; set; }

        /// <summary>
        /// Initializes a new instance of the <see cref="DbServiceContext"/> class.
        /// </summary>
        /// <param name="connectString">The connect string.</param>
        public DbServiceContext(string connectString)
        {
            ConnectString = connectString;
        }

        /// <summary>
        /// Initializes a new instance of the <see cref="DbServiceContext"/> class.
        /// </summary>
        /// <param name="options">The options.</param>
        public DbServiceContext(DbContextOptions<DbServiceContext> options) : base(options)
        {
            // just empty
        }

        /// <summary>
        /// <para>
        /// Override this method to configure the database (and other options) to be used for this context.
        /// This method is called for each instance of the context that is created.
        /// The base implementation does nothing.
        /// </para>
        /// <para>
        /// In situations where an instance of <see cref="T:Microsoft.EntityFrameworkCore.DbContextOptions" /> may or may not have been passed
        /// to the constructor, you can use <see cref="P:Microsoft.EntityFrameworkCore.DbContextOptionsBuilder.IsConfigured" /> to determine if
        /// the options have already been set, and skip some or all of the logic in
        /// <see cref="M:Microsoft.EntityFrameworkCore.DbContext.OnConfiguring(Microsoft.EntityFrameworkCore.DbContextOptionsBuilder)" />.
        /// </para>
        /// </summary>
        /// <param name="optionsBuilder">A builder used to create or modify options for this context. Databases (and other extensions)
        /// typically define extension methods on this object that allow you to configure the context.</param>
        protected override void OnConfiguring(DbContextOptionsBuilder optionsBuilder)
        {
            optionsBuilder.UseSqlServer(ConnectString);
        }

        /// <summary>
        /// Override this method to further configure the model that was discovered by convention from the entity types
        /// exposed in <see cref="T:Microsoft.EntityFrameworkCore.DbSet`1" /> properties on your derived context. The resulting model may be cached
        /// and re-used for subsequent instances of your derived context.
        /// </summary>
        /// <param name="modelBuilder">The builder being used to construct the model for this context. Databases (and other extensions) typically
        /// define extension methods on this object that allow you to configure aspects of the model that are specific
        /// to a given database.</param>
        /// <remarks>
        /// If a model is explicitly set on the options for this context (via <see cref="M:Microsoft.EntityFrameworkCore.DbContextOptionsBuilder.UseModel(Microsoft.EntityFrameworkCore.Metadata.IModel)" />)
        /// then this method will not be run.
        /// </remarks>
        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            modelBuilder.Entity<User>()
                .HasIndex(u => u.UserName)
                .IsUnique();

            modelBuilder.Entity<User>()
                .HasIndex(u => u.UniqueExternalUserId)
                .IsUnique();
        }

        /// <summary>
        /// Gets or sets the issues.
        /// </summary>
        /// <value>
        /// The issues.
        /// </value>
        public DbSet<Issue> Issues { get; set; }

        /// <summary>
        /// Gets or sets the suggestions.
        /// </summary>
        /// <value>
        /// The suggestions.
        /// </value>
        public DbSet<Suggestion> Suggestions { get; set; }

        /// <summary>
        /// Gets or sets the user.
        /// </summary>
        /// <value>
        /// The user.
        /// </value>
        // TODO: change to users
        public DbSet<User> User { get; set; }

        /// <summary>
        /// Gets or sets the staked suggestions.
        /// </summary>
        /// <value>
        /// The staked suggestions.
        /// </value>
        public DbSet<StakedSuggestion> StakedSuggestions { get; set; }

        /// <summary>
        /// Gets or sets the wallets.
        /// </summary>
        /// <value>
        /// The wallets.
        /// </value>
        public DbSet<Wallet> Wallets { get; set; }

        /// <summary>
        /// Gets or sets the wallet transactions.
        /// </summary>
        /// <value>
        /// The wallet transactions.
        /// </value>
        public DbSet<WalletTransaction> WalletTransactions { get; set; }

        /// <summary>
        /// Gets or sets the transaction types.
        /// </summary>
        /// <value>
        /// The transaction types.
        /// </value>
        public DbSet<TransactionType> TransactionTypes { get; set; }

        /// <summary>
        /// Gets or sets the votes.
        /// </summary>
        /// <value>
        /// The votes.
        /// </value>
        public DbSet<Vote> Votes { get; set; }
    }
}
