using System;
using System.Collections.Generic;
using System.Diagnostics.CodeAnalysis;
using System.Linq;
using System.Runtime.Serialization;
using System.Text.Json;
using System.Threading.Tasks;

namespace Pnyx.ApiClient.Solana.User
{
    [Serializable]
    public class PnyxAccount : ISerializable, IEqualityComparer<PnyxAccount>
    {
        #region Statics

        public static PnyxAccount Create(string userName)
        {
            return new PnyxAccount(userName);
        }

        #endregion

        #region Construction / Initialization

        public PnyxAccount(string userName)
        {
            UserName = userName;
        }

        protected PnyxAccount(SerializationInfo info, StreamingContext context)
        {
            if (info == null)
            {
                throw new ArgumentNullException("info");
            }
            UserName = info.GetString("UserName");
        }

        #endregion

        #region Properties

        public string UserName { get; }

        #endregion

        #region Public Methods

        #region ISerializable

        public void GetObjectData(SerializationInfo info, StreamingContext context)
        {
            if (info == null)
            {
                throw new ArgumentNullException("info");
            }
            info.AddValue("UserName", UserName);
        }

        #region IEqualityComparer<PnyxAccount>

        public bool Equals(PnyxAccount? x, PnyxAccount? y)
        {
            return x.UserName.Equals(y.UserName);
        }

        public int GetHashCode([DisallowNull] PnyxAccount obj)
        {
            return obj.UserName.GetHashCode();
        }

        #endregion

        #endregion

        public void Open()
        {

        }

        public void Close()
        {

        }

        #endregion
    }
}
