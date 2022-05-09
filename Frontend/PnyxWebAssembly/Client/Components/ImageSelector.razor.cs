using System;
using System.Collections.Generic;
using System.IO;
using Microsoft.AspNetCore.Components;
using Microsoft.AspNetCore.Components.Web;
using MudBlazor;
using PnyxWebAssembly.Client.Entities;

namespace PnyxWebAssembly.Client.Components
{
    /// <summary>
    /// Implementation of the image selector
    /// </summary>
    /// <seealso cref="ComponentBase" />
    public partial class ImageSelector
    {
        /// <summary>
        /// Gets or sets the mud dialog.
        /// </summary>
        /// <value>
        /// The mud dialog.
        /// </value>
        [CascadingParameter] 
        private MudDialogInstance MudDialog { get; set; }

        [Inject]
        private IDialogService DialogService { get; set; }

        /// <summary>
        /// Gets or sets the snackbar.
        /// </summary>
        /// <value>
        /// The snackbar.
        /// </value>
        [Inject]
        private ISnackbar Snackbar { get; set; }

        /// <summary>
        /// Gets or sets the hashtags.
        /// </summary>
        /// <value>
        /// The hashtags.
        /// </value>
        [Parameter]
        public IEnumerable<HashtagSelected> Hashtags { get; set; }

        /// <summary>
        /// Gets or sets the search text.
        /// </summary>
        /// <value>
        /// The search text.
        /// </value>
        public string SearchText { get; set; }

        /// <summary>
        /// Gets the image items.
        /// </summary>
        /// <value>
        /// The image items.
        /// </value>
        public List<ImageItem> ImageItems { get; } = new();

        /// <summary>
        /// Gets or sets the selected paper.
        /// </summary>
        /// <value>
        /// The selected paper.
        /// </value>
        private int SelectedPaper { get; set; }

        /// <summary>
        /// Gets or sets the form.
        /// </summary>
        /// <value>
        /// The form.
        /// </value>
        private MudForm Form { get; set; }

        private string[] Errors { get; set; } = { };
        
        public ICollection<object> SelectedValues { get; set; }

        /// <summary>
        /// Gets or sets a value indicating whether [keep search text].
        /// </summary>
        /// <value>
        ///   <c>true</c> if [keep search text]; otherwise, <c>false</c>.
        /// </value>
        private bool KeepSearchText { get; set; }

        /// <summary>
        /// Submits this instance.
        /// </summary>
        private void Submit()
        {
            MudDialog.Close(DialogResult.Ok(SelectedPaper));
        }

        /// <summary>
        /// Cancels this instance.
        /// </summary>
        private void Cancel() => MudDialog.Cancel();

        /// <summary>
        /// Called when [search click].
        /// </summary>
        private void OnSearchClick()
        {
            InvokeSearch();
        }

        /// <summary>
        /// Invokes the search.
        /// </summary>
        private async void InvokeSearch()
        {
            bool searchTextHasChip = false;

            foreach (HashtagSelected hashtagSelected in Hashtags)
            {
                if (SearchText.Contains(hashtagSelected.Hashtag[1..], StringComparison.OrdinalIgnoreCase))
                {
                    searchTextHasChip = true;
                    break;
                }
            }

            if (!searchTextHasChip)
            {
                SelectedValues = null;
                KeepSearchText = true;
            }

            ImageItems.Clear();

            Snackbar.Clear();
            Snackbar.Configuration.PositionClass = Defaults.Classes.Position.TopCenter;
            Snackbar.Add($"Searching: {SearchText}");

            // TODO: check temp folder for pixabay api key file
            // TODO: ask user in dialog for pixabay api key if not found
            // TODO: use temp folder for docker as volume to retain the pixabay api key

            string tempFolder = Path.GetTempPath();

            string fullPathToPixabayApiKeyFile = Path.Combine(tempFolder, "PixabayApiKey.txt");

            if (!File.Exists(fullPathToPixabayApiKeyFile))
            {
                DialogParameters parameters = new DialogParameters
                {
                    ["infotext"] = "Der Pixabay API-Key ist für diesen Test erforderlich und muss einmal eingegeben werden",
                    ["label"] = "Pixabay API-Key"
                };

                var options = new DialogOptions
                {
                    Position = DialogPosition.Center
                };

                IDialogReference dialog = DialogService.Show<ApiKeyDialog>("Pixabay API-Key", parameters, options);
                DialogResult result = await dialog.Result;

                if (result.Cancelled)
                {
                    MudDialog.Cancel();
                    return;
                }

                string apiKey = result.Data.ToString();

                // TODO: check api key - if it is wrong display dialog again and inform user that the api key is invalid

                // TODO: save api key to temp file if the api key is correct
            }

            // TODO: load api key

            // TODO: do query

            // TODO: check api key - if it is wrong display dialog again and inform user that the api key is invalid

            // TODO: save api key to temp file if the api key is correct

            // TODO: link results to image items

            // TODO: display error message if no image was found

            for (int i = 0; i < 96; i++)
            {
                ImageItems.Add(new ImageItem
                {
                    Id = i + 1
                });
            }

            // TODO: another MudBlazor error - MudChipSet is not re-rendered but must be - because it is changed!
            await InvokeAsync(StateHasChanged);
        }

        /// <summary>
        /// Called when [selected chip changed].
        /// </summary>
        /// <param name="mudChip">The mud chip.</param>
        private void OnSelectedChipChanged(MudChip mudChip)
        {
            Form.ResetValidation();

            if (mudChip == null)
            {
                ImageItems.Clear();

                if (!KeepSearchText)
                {
                    SearchText = string.Empty;
                }
                else
                {
                    InvokeSearch();
                    KeepSearchText = false;
                }

                return;
            }

            SearchText = mudChip.Text.Replace("#", string.Empty);

            InvokeSearch();
        }

        /// <summary>
        /// Raises the <see cref="E:SearchKeyUp" /> event.
        /// </summary>
        /// <param name="e">The <see cref="KeyboardEventArgs"/> instance containing the event data.</param>
        private void OnSearchKeyUp(KeyboardEventArgs e)
        {
            if (e.Code == "Enter" && 
                !string.IsNullOrEmpty(SearchText) &&
                SearchText.Length >= 3)
            {
                Form.ResetValidation();
                InvokeSearch();
            }
        }

        /// <summary>
        /// Validates the search.
        /// </summary>
        /// <param name="searchText">The search text.</param>
        /// <returns></returns>
        private IEnumerable<string> ValidateSearch(string searchText)
        {
            if (string.IsNullOrWhiteSpace(searchText))
            {
                yield return "Ein Suchtext muss angegeben werden";
                yield break;
            }

            if (searchText.Length < 3)
            {
                yield return "Der Suchtext muss mindestens 3 Zeichen lang sein";
            }
        }
    }
}
