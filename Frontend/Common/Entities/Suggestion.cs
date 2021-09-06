using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the suggestion
    /// </summary>
    [Table("Suggestions")]
    public class Suggestion
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="Suggestion"/> class.
        /// </summary>
        public Suggestion()
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
        public Guid Id { get; set; }

        /// <summary>
        /// Gets or sets the short description.
        /// </summary>
        /// <value>
        /// The short description.
        /// </value>
        public string ShortDescription { get; set; }

        /// <summary>
        /// Gets or sets the description.
        /// </summary>
        /// <value>
        /// The description.
        /// </value>
        public string Description { get; set; }

        /// <summary>
        /// Gets or sets the create date.
        /// </summary>
        /// <value>
        /// The create date.
        /// </value>
        [DatabaseGenerated(DatabaseGeneratedOption.Computed)]
        public DateTime CreateDate { get; set; }

        /// <summary>
        /// Gets or sets the stake count.
        /// </summary>
        /// <value>
        /// The stake count.
        /// </value>
        public int StakeCount { get; set; }

        /// <summary>
        /// Gets or sets the stake area.
        /// </summary>
        /// <value>
        /// The stake area.
        /// </value>
        public StakeArea StakeArea { get; set; }
    }
}
