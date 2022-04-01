using System.Collections.Generic;
using Microsoft.AspNetCore.Components;
using Microsoft.Extensions.Options;
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
        public string Hashtags { get; set; }

        private int SelectedPaper { get; set; }

        public List<ImageItem> ImageItems { get; } = new List<ImageItem>();

        public ImageSelector()
        {
            for (int i = 0; i < 96; i++)
            {
                ImageItems.Add(new ImageItem
                {
                    Id = i
                });
            }
        }

        private void Submit()
        {
            MudDialog.Close(DialogResult.Ok(true));

            Snackbar.Clear();
            Snackbar.Configuration.PositionClass = Defaults.Classes.Position.TopCenter;
            Snackbar.Add($"Selected Item: {SelectedPaper}");
        }

        private void Cancel() => MudDialog.Cancel();
    }
}
