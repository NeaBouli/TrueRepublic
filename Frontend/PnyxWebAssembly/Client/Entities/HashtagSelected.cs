namespace PnyxWebAssembly.Client.Entities
{
    /// <summary>
    /// Implementation of the hashtag selected
    /// </summary>
    public class HashtagSelected
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="HashtagSelected"/> class.
        /// </summary>
        /// <param name="hashtag">The hashtag.</param>
        /// <param name="selected">if set to <c>true</c> [selected].</param>
        public HashtagSelected(string hashtag, bool selected)
        {
            Hashtag = hashtag;
            Selected = selected;
        }

        /// <summary>
        /// Gets or sets the hashtag.
        /// </summary>
        /// <value>
        /// The hashtag.
        /// </value>
        public string Hashtag { get; set; }

        /// <summary>
        /// Gets or sets a value indicating whether this <see cref="HashtagSelected"/> is selected.
        /// </summary>
        /// <value>
        ///   <c>true</c> if selected; otherwise, <c>false</c>.
        /// </value>
        public bool Selected { get; set; }
    }
}
