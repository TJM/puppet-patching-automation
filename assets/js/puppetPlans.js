var puppetEnvironments = new Bloodhound({
  datumTokenizer: Bloodhound.tokenizers.whitespace,
  queryTokenizer: Bloodhound.tokenizers.whitespace,
  prefetch: {
    url: '/config/puppetServer/' + puppetServerID + '/environments',
    transform: function(r) { return r.environments.map(function(env) { return env['name']; }) },
  },
});

var puppetPlans = new Bloodhound({
  datumTokenizer: Bloodhound.tokenizers.nonword,
  queryTokenizer: Bloodhound.tokenizers.nonword,
  prefetch: {
    url: getPuppetPlansURL(),
    transform: function(r) { return r.plans.items.map(function(item) { return item['name']; }) },
  },
});

function getPuppetPlansURL() {
  env = $('#PuppetEnvironment').val()
  if (env === '') { // default to production
    env = 'production'
  }
  return '/config/puppetServer/' + puppetServerID + '/apiPlans/' + env
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


function refreshPuppetPlans() {
  console.log('Refreshing puppetPlans...')
  $('#PuppetPlans-Refresh').addClass('fa-spin')
  puppetPlans.clear();
  puppetPlans.clearPrefetchCache();
  puppetPlans.initialize(true)
    .done(function() {
      $('#PuppetPlans-Refresh').removeClass('fa-spin')
      console.log('Successfully Refreshed puppetPlans');
    })
    .fail(function() {
      $('#PuppetPlans-Refresh').removeClass('fa-spin').addClass('fa-red')
      console.log('ERROR Refreshing puppetPlans');
    });
}

$(document).ready(function(){

  // Refresh Buttons
  $("#PuppetEnvironment-Refresh").click(refreshPuppetEnvironments)
  $("#PuppetPlans-Refresh").click(refreshPuppetPlans)
  $("#PuppetEnvironment").change(function() {
    if ($("#PuppetEnvironment").val() === '') {
      alert("Puppet Environment is required")
    } else {
      console.log("Environment was changed, updating plans...");
      puppetPlans.prefetch.url = getPuppetPlansURL();
      refreshPuppetPlans();
    }
  })

  // Typeahead for PuppetEnvironment
  $('#PuppetEnvironmentDiv .typeahead').typeahead(null, {
    name: 'puppetEnvironments',
    source: puppetEnvironments,
    limit: 5,
  });

  // Typeahead for PuppetPlan
  $('#PuppetPlansDiv .typeahead').typeahead(null, {
    name: 'puppetPlans',
    source: puppetPlans,
    limit: 10,
  });

})
