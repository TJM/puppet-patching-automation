
$(document).ready(function(){

  // Enable ToolTips
  $('[data-toggle="tooltip"]').tooltip();

  // PatchServer Button
  $("button.patchButton").click(function(){
    // spin?
    $.post($(this).val(), // URL
    {}, // DATA
    function(data, status){ // CALLBACK
      console.log(data)
      var win = window.open(data.link, '_blank');
      if (win) {
          //Browser has allowed it to be opened
          win.focus();
      } else {
          //Browser has blocked it
          alert('Please allow popups for this website');
      }
      // Stop Spin?
      $(this).prop("disabled", true)
    }),
    "json" // dataType
  });

});
