using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Pnyx.ApiClient.Core.Accounts
{
    public enum PnyxAccountRoles
    {
        None = 0,
        Administrator = 1 << 1,
        AccountManager = 1 << 2,
        Proposer = 1 << 3,
        Voter = 1 << 4,
        All = (1 << 5) - 1
    }

    public class PnyxAccount
    {
        public PnyxAccount()
        {
            ID = Guid.NewGuid().ToString().ToUpper();
        }

        public string ID { get; }
    }
}
