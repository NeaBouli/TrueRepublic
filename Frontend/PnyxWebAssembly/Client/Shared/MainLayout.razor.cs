using System.Threading.Tasks;
using Microsoft.AspNetCore.Components;
using Microsoft.AspNetCore.Components.WebAssembly.Authentication;

namespace PnyxWebAssembly.Client.Shared
{
    /// <summary>
    /// Implementation of the MainLayout
    /// </summary>
    /// <seealso cref="Microsoft.AspNetCore.Components.LayoutComponentBase" />
    public partial class MainLayout : LayoutComponentBase
    {
        /// <summary>
        /// Gets or sets the navigation manager.
        /// </summary>
        /// <value>
        /// The navigation manager.
        /// </value>
        [Inject]
        private NavigationManager NavigationManager { get; set; }

        /// <summary>
        /// Gets or sets the sign out session state manager.
        /// </summary>
        /// <value>
        /// The sign out session state manager.
        /// </value>
        [Inject]
        private SignOutSessionStateManager SignOutSessionStateManager { get; set; }

        /// <summary>
        /// Gets or sets the total balance.
        /// </summary>
        /// <value>
        /// The total balance.
        /// </value>
        private int _totalBalance;

        /// <summary>
        /// Gets or sets the total balance.
        /// </summary>
        /// <value>
        /// The total balance.
        /// </value>
        public int TotalBalance
        {
            get => _totalBalance;
            set
            {
                _totalBalance = value;
                InvokeAsync(StateHasChanged);
            }
        }

        private string _userName;
        private string _avatarImage;

        /// <summary>
        /// Gets or sets the name of the user.
        /// </summary>
        /// <value>
        /// The name of the user.
        /// </value>
        public string UserName
        {
            get => _userName;
            set
            {
                _userName = value;
                InvokeAsync(StateHasChanged);
            }
        }

        /// <summary>
        /// Gets the avatar image.
        /// </summary>
        /// <value>
        /// The avatar image.
        /// </value>
        public string AvatarImage
        {
            get => _avatarImage;
            set
            {
                _avatarImage = value;
                InvokeAsync(StateHasChanged);
            }
        }

        /// <summary>
        /// Gets the name of the avatar.
        /// </summary>
        /// <value>
        /// The name of the avatar.
        /// </value>
        public string AvatarName => string.IsNullOrEmpty(_userName) ? string.Empty : _userName.ToUpperInvariant().Substring(0, 1);

        /// <summary>
        /// Begins the sign out.
        /// </summary>
        private async Task BeginSignOut()
        {
            await SignOutSessionStateManager.SetSignOutState();
            NavigationManager.NavigateTo("authentication/logout");
        }

        /// <summary>
        /// Begins the sign in.
        /// </summary>
        private void BeginSignIn()
        {
            NavigationManager.NavigateTo("authentication/login");
        }

        /// <summary>
        /// Begins the register.
        /// </summary>
        private void BeginRegister()
        {
            NavigationManager.NavigateTo("authentication/register");
        }

        /// <summary>
        /// Goto the profile.
        /// </summary>
        private void GotoProfile()
        {
            NavigationManager.NavigateTo("authentication/profile");
        }
    }
}
