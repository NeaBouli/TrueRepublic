using System.ComponentModel.DataAnnotations;

namespace PnyxWebAssembly.Client.Entities
{
    /// <summary>
    /// Implementation ot the add user wizard form
    /// </summary>
    public class AddUserWizardForm
    {
        [Required]
        [StringLength(15, ErrorMessage = "Username must be between 5 and 15 characters long", MinimumLength = 5)]
        public string UserName { get; set; }
    }
}
