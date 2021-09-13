namespace Common.Entities
{
    /// <summary>
    /// The transaction type names
    /// </summary>
    public enum TransactionTypeNames
    {
        /// <summary>
        /// The genesis transaction
        /// </summary>
        Genesis,

        /// <summary>
        /// The add item
        /// </summary>
        AddIssue,

        /// <summary>
        /// The add suggestion
        /// </summary>
        AddSuggestion,

        /// <summary>
        /// The stake suggestion
        /// </summary>
        StakeSuggestion,

        /// <summary>
        /// The stake suggestion rollback
        /// </summary>
        StakeSuggestionRollback
    }
}
