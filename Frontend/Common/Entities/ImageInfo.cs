using System;
using System.ComponentModel.DataAnnotations;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the image info
    /// </summary>
    public class ImageInfo
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ImageInfo"/> class.
        /// </summary>
        public ImageInfo()
        {
            Id = Guid.NewGuid();
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
        /// Gets the hashtags.
        /// </summary>
        /// <value>
        /// The hashtags.
        /// </value>
        [Required]
        public string Hashtags { get; set; }

        /// <summary>
        /// Gets or sets the filename.
        /// </summary>
        /// <value>
        /// The filename.
        /// </value>
        [Required]
        public string Filename { get; set; }
    }
}
