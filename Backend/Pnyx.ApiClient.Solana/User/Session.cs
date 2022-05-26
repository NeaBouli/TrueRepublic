using Pnyx.ApiClient.Interfaces.Session;


namespace Pnyx.ApiClient.Solana.User
{
    public class Session : ISession
    {
        public static Session Create(IUser user)
        {
            return new Session();
        }
    }
}
