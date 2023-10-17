var patchWindows = new Bloodhound({
  datumTokenizer: function(d) { return Bloodhound.tokenizers.whitespace(d.value); },
  queryTokenizer: Bloodhound.tokenizers.whitespace,
  identify: function(obj) { return obj.value; },
  prefetch: '/data/patchWindows',
});

function refreshPatchWindows() {
  $('#PatchWindow-Refresh').addClass('fa-spin')
  patchWindows.clear();
  patchWindows.clearPrefetchCache();
  promise=patchWindows.initialize(true);
  promise
    .done(function() {
      $('#PatchWindow-Refresh').removeClass('fa-spin')
      console.log('Successfully Refreshed PatchWindows');
    })
    .fail(function() {
      $('#PatchWindow-Refresh').removeClass('fa-spin').addClass('fa-red')
      console.log('ERROR Refreshing PatchWindows');
    });
}

function searchPatchWindows(q, sync) {
  if (q === '') {
    sync(patchWindows.all());
  } else {
    patchWindows.search(q, sync);
  }
}

$(document).ready(function(){

  // PatchServer Button
  $("i#PatchWindow-Refresh").click(refreshPatchWindows)

  // Typeahead for PatchWindow
  $('#PatchWindow .typeahead').typeahead({
    minLength: 0,
  }, {
    name: 'PatchWindows',
    source: searchPatchWindows,
    limit: 10,
    display: 'value',
    templates: {
      suggestion: function (data) {
        return '<div><strong>' + data.value + '</strong> - ' + data.count + ' hosts</div>';
      }
    },
  })
  .on('typeahead:asyncrequest', function() {
    console.log("Loading...")
    $('.Typeahead-spinner').show();
  })
  .on('typeahead:asynccancel typeahead:asyncreceive', function() {
    console.log("Done.")
    $('.Typeahead-spinner').hide();
  })
  .on('typeahead:initializing', function() {
    console.log("Initializing...")
    $('.Typeahead-spinner').show();
  })
  .on('typeahead:done', function() {
    console.log("Done.")
    $('.Typeahead-spinner').hide();
  });

})
