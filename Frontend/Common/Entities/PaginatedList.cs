namespace Common.Entities
{
    public class PaginatedList
    {
        /// <summary>
        /// Gets or sets the items per page.
        /// </summary>
        /// <value>
        /// The items per page.
        /// </value>
        public int ItemsPerPage { get; set; }

        /// <summary>
        /// Gets or sets the page.
        /// </summary>
        /// <value>
        /// The page.
        /// </value>
        public int Page { get; set; }

        /// <summary>
        /// Gets or sets a value indicating whether [get details].
        /// </summary>
        /// <value>
        ///   <c>true</c> if [get details]; otherwise, <c>false</c>.
        /// </value>
        public bool GetDetails { get; set; }

        /// <summary>
        /// Gets the skip.
        /// </summary>
        /// <value>
        /// The skip.
        /// </value>
        public int Skip => (Page - 1) * ItemsPerPage;
    }
}
