using System;
using System.Collections.Generic;
using System.Data;
using Common.Entities;

namespace Common.Interfaces
{
    public interface IUserService
    {
        /// <summary>
        /// Gets the users.
        /// </summary>
        /// <returns></returns>
        List<User> GetUsers();

        /// <summary>
        /// Gets the user by identifier.
        /// </summary>
        /// <param name="id">The identifier.</param>
        /// <returns>The user if found otherwise null</returns>
        User GetUserById(Guid id);

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>
        /// The number of imported records
        /// </returns>
        int Import(DataTable dataTable);

        /// <summary>
        /// Adds the specified user.
        /// </summary>
        /// <param name="user">The user.</param>
        void Add(User user);
    }
}