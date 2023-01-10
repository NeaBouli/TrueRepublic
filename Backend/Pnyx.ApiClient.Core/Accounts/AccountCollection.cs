using Pnyx.ApiClient.Core.Accounts;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Pnyx.SmartContracts.Solidity.Contracts.Accounts
{
    public class AccountCollection<AccountType> where AccountType : PnyxAccount
    {
        private static Dictionary<string, AccountType> _accounts = new Dictionary<string, AccountType>();

        public bool AddAccount(AccountType account)
        {
            bool containsKey = _accounts.ContainsKey(account.ID);
            if (!containsKey)
            {
                _accounts[account.ID] = account;
            }
            return containsKey;
        }

        public AccountType GetAccount(string accountID)
        {
            if (_accounts.ContainsKey(accountID))
            {
                return _accounts[accountID];
            }
            return null;
        }

        public bool RemoveAccount(AccountType account)
        {
            return _accounts.Remove(account.ID);
        }
    }
}
