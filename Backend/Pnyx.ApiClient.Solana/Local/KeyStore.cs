using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Runtime.Serialization;
using System.Runtime.Serialization.Formatters.Binary;
using System.Runtime.Serialization.Json;
using System.Threading.Tasks;
using Solnet.KeyStore;
using Pnyx.ApiClient.Solana.User;

namespace Pnyx.ApiClient.Solana.Local
{
    public class KeyStore
    {
        #region Statics

        public static KeyStore Instance;

        static KeyStore()
        {
            Instance = new KeyStore();
        }

        #endregion

        #region Class Members

        private SecretKeyStoreService _keyStoreService;

        #endregion

        #region Construction / Initialization

        private KeyStore()
        {
            _keyStoreService = new SecretKeyStoreService();
        }

        #region Public Methods

        public string Encrypt(PnyxAccount pnyxAccount, string password)
        {
            string jsonString = null;
            IFormatter formatter = new BinaryFormatter();
            using (MemoryStream memStream = new MemoryStream())
            {
                formatter.Serialize(memStream, pnyxAccount);
                byte[] data = memStream.GetBuffer();
                jsonString = _keyStoreService.EncryptAndGenerateDefaultKeyStoreAsJson(password, data, "");
            }
            return jsonString;
        }

        public void EncryptToFile(PnyxAccount pnyxAccount, string password)
        {
            IFormatter formatter = new BinaryFormatter();
            using (MemoryStream memStream = new MemoryStream())
            {
                formatter.Serialize(memStream, pnyxAccount);
                byte[] data = memStream.GetBuffer();
                //_keyStoreService.
            }
        }

        public void DecryptFromFile(PnyxAccount pnyxAccount, string password)
        {
            //_keyStoreService
        }

        public PnyxAccount Decrypt(string jsonString, string password)
        {
            PnyxAccount pnyxAccount = null;
            byte[] data = _keyStoreService.DecryptKeyStoreFromJson(password, jsonString);
            IFormatter formatter = new BinaryFormatter();
            using (MemoryStream memStream = new MemoryStream(data))
            {
                pnyxAccount = (PnyxAccount)formatter.Deserialize(memStream);
            }
            return pnyxAccount;
        }

        #endregion

        #endregion
    }
}
