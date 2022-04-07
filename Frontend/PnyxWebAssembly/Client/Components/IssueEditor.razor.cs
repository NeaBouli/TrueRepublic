using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Reflection.Metadata;
using System.Security.Claims;
using System.Threading.Tasks;
using Common.Entities;
using Microsoft.AspNetCore.Components;
using Microsoft.AspNetCore.Components.Authorization;
using MudBlazor;
using PnyxWebAssembly.Client.Entities;
using PnyxWebAssembly.Client.Services;

namespace PnyxWebAssembly.Client.Components
{
    /// <summary>
    /// Implementation of the issue editor
    /// </summary>
    /// <seealso cref="Microsoft.AspNetCore.Components.ComponentBase" />
    public partial class IssueEditor
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
        /// Gets or sets the authentication state provider.
        /// </summary>
        /// <value>
        /// The authentication state provider.
        /// </value>
        [Inject]
        private AuthenticationStateProvider AuthenticationStateProvider { get; set; }

        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        [Inject]
        private IHttpClientFactory ClientFactory { get; set; }

        [Inject]
        private IDialogService DialogService { get; set; }

        [Inject]
        private ISnackbar Snackbar { get; set; }

        /// <summary>
        /// Gets or sets the action.
        /// </summary>
        /// <value>
        /// The action.
        /// </value>
        [Parameter]
        public string Action { get; set; }

        private bool Success { get; set; }

        private string[] Errors { get; set; } = { };

        private IssueService IssueService { get; set; }

        private Issue Issue { get; set; }

        private string IssueImage { get; set; }

        private EditMode EditMode { get; set; }

        private User User { get; set; }

        private string ErrorMessage { get; set; }

        /// <summary>
        /// Gets or sets a value indicating whether [show hashtag chips].
        /// </summary>
        /// <value>
        ///   <c>true</c> if [show hashtag chips]; otherwise, <c>false</c>.
        /// </value>
        private bool ShowHashtagChips { get; set; }

        /// <summary>
        /// Gets or sets the hashtag value.
        /// </summary>
        /// <value>
        /// The hashtag value.
        /// </value>
        private string HashtagValue { get; set; }

        /// <summary>
        /// Method invoked when the component is ready to start, having received its
        /// initial parameters from its parent in the render tree.
        /// Override this method if you will perform an asynchronous operation and
        /// want the component to refresh when that operation is completed.
        /// </summary>
        protected override async Task OnInitializedAsync()
        {
            Issue = new Issue
            {
                Tags = string.Empty
            };

            IssueService = new IssueService();
            IssueService.ClientFactory = ClientFactory;

            AuthenticationState authState = await AuthenticationStateProvider.GetAuthenticationStateAsync();
            ClaimsPrincipal user = authState.User;

            string userEMail = user.Identity?.Name;

            if (!string.IsNullOrEmpty(userEMail))
            {
                List<Claim> claims = user.Claims.ToList();

                if (claims.Count >= 3)
                {
                    string externalUserId = claims[2].Value;

                    if (UserService.User != null && 
                        UserService.User.UniqueExternalUserId.ToString() == externalUserId)
                    {
                        User = UserService.User;
                    }
                }
            }

            if (Action.Equals("add", StringComparison.OrdinalIgnoreCase))
            {
                if (User != null)
                {
                    EditMode = EditMode.AddNew;
                }
                else
                {
                    ErrorMessage = "Error: user has no permission to add issue";
                    return;
                }
            }
            else
            {
                if (!Guid.TryParse(Action, out Guid issueId))
                {
                    ErrorMessage = "Error: given parameter is no valid issue id";
                    return;
                }

                HttpClient client;

                if (User != null)
                {
                    client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Private");
                    EditMode = EditMode.Edit;
                }
                else
                {
                    client = ClientFactory.CreateClient("PnyxWebAssembly.ServerAPI.Public");
                    EditMode = EditMode.ReadOnly;
                }

                // TODO: load issue with all proposals
            }
        }

        /// <summary>
        /// Adds the hashtag click.
        /// </summary>
        private async Task AddHashtagClick()
        {
            await AddHashtag();
        }

        /// <summary>
        /// Adds the hashtag.
        /// </summary>
        private async Task AddHashtag()
        {
            ShowHashtagChips = false;

            if (string.IsNullOrEmpty(HashtagValue))
            {
                return;
            }

            ShowHashtagChips = true;

            if (!HashtagValue.StartsWith("#"))
            {
                HashtagValue = HashtagValue[..1].ToUpper() + HashtagValue[1..];
                HashtagValue = $"#{HashtagValue}";
            }
            else
            {
                HashtagValue = HashtagValue[..2].ToUpper() + HashtagValue[2..];
            }

            List<string> hashtags = Issue.GetTags().ToList();

            if (hashtags.Contains(HashtagValue, StringComparer.OrdinalIgnoreCase))
            {
                HashtagValue = string.Empty;
                ErrorMessage = "Hashtag with the same name already added";
                return;
            }

            Issue.AddTag(HashtagValue);

            await SetImageForHashtags();

            HashtagValue = string.Empty;
        }

        private async Task SetImageForHashtags()
        {
            KeyValuePair<string, string>? result = await IssueImageService.GetImageFromServiceForHashtags(Issue.Tags);

            if (result != null)
            {
                Issue.ImageName = ((KeyValuePair<string, string>) result).Key;
                IssueImage = ((KeyValuePair<string, string>) result).Value;
            }
        }

        /// <summary>Called when [mud chip close].</summary>
        /// <param name="chip">The chip.</param>
        private async Task OnMudChipClose(MudChip chip)
        {
            Issue.RemoveTag(chip.Text);

            if (Issue.Tags.Length == 0)
            {
                ShowHashtagChips = false;
                Issue.ImageName = string.Empty;
                IssueImage = string.Empty;
            }
            else
            {
                await SetImageForHashtags();
            }
        }

        /// <summary>
        /// Search for hashtags
        /// </summary>
        /// <param name="value">The value to search for</param>
        /// <returns>An enumeration of all found hashtags</returns>
        private Task<IEnumerable<string>> SearchHashtags(string value)
        {
            return IssueService.GetHashtags(value);
        }

        private async Task EditImageClick()
        {
            List<string> hashtags = Issue.GetTags().ToList();

            List<HashtagSelected> hashtagsSelected = new List<HashtagSelected>();

            for (var i = 0; i < hashtags.Count; i++)
            {
                hashtagsSelected.Add(new HashtagSelected(hashtags[i], i == 0));
            }

            DialogParameters parameters = new DialogParameters { ["hashtags"] = hashtagsSelected };

            var options = new DialogOptions
            {
                MaxWidth = MaxWidth.Large,
                FullWidth = true,
                Position = DialogPosition.TopCenter
            };

            IDialogReference dialog = DialogService.Show<ImageSelector>("Edit Image", parameters, options);
            DialogResult result = await dialog.Result;

            Snackbar.Clear();
            Snackbar.Configuration.PositionClass = Defaults.Classes.Position.TopCenter;
            Snackbar.Add($"Selected Item: {result.Data}");
        }

        private async Task OnValueChanged(string arg)
        {
            await AddHashtag();
        }
    }
}
