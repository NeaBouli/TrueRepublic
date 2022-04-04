using System.Collections.Generic;
using Microsoft.AspNetCore.Components;
using Microsoft.AspNetCore.Components.Web;
using MudBlazor;

namespace PnyxWebAssembly.Client.Components
{
    public class ImageItem
    {
        public int Id { get; set; }
    }

    public partial class ImageSelector
    {
        [CascadingParameter] 
        private MudDialogInstance MudDialog { get; set; }

        [Inject]
        private ISnackbar Snackbar { get; set; }

        [Parameter]
        public IEnumerable<string> Hashtags { get; set; }

        public string SearchText { get; set; }

        public List<ImageItem> ImageItems { get; } = new();

        private int _selectedPaper;

        private MudForm _form;

        private string[] _errors = { };

        private void Submit()
        {
            MudDialog.Close(DialogResult.Ok(_selectedPaper));
        }

        private void Cancel() => MudDialog.Cancel();

        private void OnSearchClick()
        {
            InvokeSearch();
        }

        private void InvokeSearch()
        {
            ImageItems.Clear();

            Snackbar.Clear();
            Snackbar.Configuration.PositionClass = Defaults.Classes.Position.TopCenter;
            Snackbar.Add($"Searching: {SearchText}");

            for (int i = 0; i < 96; i++)
            {
                ImageItems.Add(new ImageItem
                {
                    Id = i + 1
                });
            }
        }

        private void OnSelectedChipChanged(MudChip mudChip)
        {
            _form.ResetValidation();

            if (mudChip == null)
            {
                ImageItems.Clear();
                SearchText = string.Empty;
                return;
            }

            SearchText = mudChip.Text.Replace("#", string.Empty);

            InvokeSearch();
        }

        private void OnSearchKeyUp(KeyboardEventArgs e)
        {
            if (e.Code == "Enter" && 
                !string.IsNullOrEmpty(SearchText) &&
                SearchText.Length >= 3)
            {
                _form.ResetValidation();
                InvokeSearch();
            }
        }

        private IEnumerable<string> ValidateSearch(string searchText)
        {
            if (string.IsNullOrWhiteSpace(searchText))
            {
                yield return "Ein Suchtext muss angegeben werden";
                yield break;
            }

            if (searchText.Length < 3)
                yield return "Der Suchtext muss mindestens 3 Zeichen lang sein";
        }
    }
}
