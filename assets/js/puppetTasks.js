var puppetEnvironments = new Bloodhound({
  datumTokenizer: Bloodhound.tokenizers.whitespace,
  queryTokenizer: Bloodhound.tokenizers.whitespace,
  prefetch: {
    url: '/config/puppetServer/' + puppetServerID + '/environments',
    transform: function(r) { return r.environments.map(function(env) { return env['name']; }) },
  },
});

var puppetTasks = new Bloodhound({
  datumTokenizer: Bloodhound.tokenizers.nonword,
  queryTokenizer: Bloodhound.tokenizers.nonword,
  prefetch: {
    url: getPuppetTasksURL(),
    transform: function(r) { return r.tasks.items.map(function(item) { return item['name']; }) },
  },
});

function getPuppetTasksURL() {
  env = $('#PuppetEnvironment').val()
  if (env === '') { // default to production
    env = 'production'
  }
  return '/config/puppetServer/' + puppetServerID + '/apiTasks/' + env
}

function refreshPuppetEnvironments() {
  console.log('Refreshing puppetEnvironments...')
  $('#PuppetEnvironment-Refresh').addClass('fa-spin')
  puppetEnvironments.clear();
  puppetEnvironments.clearPrefetchCache();
  puppetEnvironments.initialize(true)
    .done(function() {
      $('#PuppetEnvironment-Refresh').removeClass('fa-spin')
      console.log('Successfully Refreshed puppetEnvironments');
    })
    .fail(function() {
      $('#PuppetEnvironment-Refresh').removeClass('fa-spin').addClass('fa-red')
      console.log('ERROR Refreshing puppetEnvironments');
    });
}


function refreshPuppetTasks() {
  console.log('Refreshing puppetTasks...')
  $('#PuppetTasks-Refresh').addClass('fa-spin')
  puppetTasks.clear();
  puppetTasks.clearPrefetchCache();
  puppetTasks.initialize(true)
    .done(function() {
      $('#PuppetTasks-Refresh').removeClass('fa-spin')
      console.log('Successfully Refreshed puppetTasks');
    })
    .fail(function() {
      $('#PuppetTasks-Refresh').removeClass('fa-spin').addClass('fa-red')
      console.log('ERROR Refreshing puppetTasks');
    });
}

$(document).ready(function(){

  // Refresh Buttons
  $("#PuppetEnvironment-Refresh").click(refreshPuppetEnvironments)
  $("#PuppetTasks-Refresh").click(refreshPuppetTasks)
  $("#PuppetEnvironment").change(function() {
    if ($("#PuppetEnvironment").val() === '') {
      alert("Puppet Environment is required")
    } else {
      console.log("Environment was changed, updating tasks...");
      puppetTasks.prefetch.url = getPuppetTasksURL();
      refreshPuppetTasks();
    }
  })

  // Typeahead for PuppetEnvironment
  $('#PuppetEnvironmentDiv .typeahead').typeahead(null, {
    name: 'puppetEnvironments',
    source: puppetEnvironments,
    limit: 5,
  });

  // Typeahead for PuppetTask
  $('#PuppetTasksDiv .typeahead').typeahead(null, {
    name: 'puppetTasks',
    source: puppetTasks,
    limit: 10,
  });

})
