using System;

namespace Common.Entities
{
    /// <summary>
    /// Implementation of the name event args
    /// </summary>
    /// <seealso cref="System.EventArgs" />
    public class NameCountEventArgs : EventArgs
    {
        /// <summary>
        /// Gets the name.
        /// </summary>
        /// <value>
        /// The name.
        /// </value>
        public string Name { get; }

        /// <summary>
        /// Gets the count.
        /// </summary>
        /// <value>
        /// The count.
        /// </value>
        public int Count { get; }

        /// <summary>
        /// Initializes a new instance of the <see cref="NameCountEventArgs"/> class.
        /// </summary>
        /// <param name="name">The name.</param>
        public NameCountEventArgs(string name, int count)
        {
            Name = name;
            Count = count;
        }
    }
}
