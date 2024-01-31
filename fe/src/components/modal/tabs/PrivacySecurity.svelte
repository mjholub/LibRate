<script defer lang="ts">
  // @ts-ignore
  import Tags from 'svelte-tags-input';
  import { _ } from 'svelte-i18n';
  import type { PrivacySecurityPreferences } from '$stores/members/prefs';
  import { createEventDispatcher } from "svelte";
	import axios from 'axios';
  let errorMessages: string[] = [];
  let confirmMutingInstance = false;

  // TODO: actual logic to fetch and cache the known network as suggestions setting for muted instances
  const knownInstances = ["bookwyrm.social"];

  $: settingsSaved = false;

  export let memberName: string;

  // TODO: fetch the current settings
  let settings: PrivacySecurityPreferences = { 
		auto_accept_follow: true,
		locally_searchable: true,
		robots_searchable: false,
		blur_nsfw: true,
		searchable_to_federated: true,
		message_autohide_words: [],
		muted_instances: []
  };

  const dispatch = createEventDispatcher();

  // we skip reading the array properties from DOM, since we're using svelte-tags-input for these fields
  // we're using reverse order for nouns in the variable names here to avoid accidentally confusing
  // them with the props of PrivacySecurityPreferences type
  const getSettingsElements = () => {
    const federatedSearchable = document.getElementById('federated-searchable') as HTMLInputElement;
    const followAutoAccept = document.getElementById('auto-accept-follow') as HTMLInputElement;
    const searchableLocally = document.getElementById('locally-searchable') as HTMLInputElement;
    const searchableToRobots = document.getElementById('robots-serchable') as HTMLInputElement;
    const nsfwBlur = document.getElementById('blur-nsfw') as HTMLInputElement;

    return { federatedSearchable, followAutoAccept, searchableLocally, searchableToRobots, nsfwBlur };
  }

  // NOTE: only making a request here (for now). Theme is saved in localStorage
  const csrfToken = document.cookie
				.split('; ')
				.find((row) => row.startsWith('csrf_'))
				?.split('=')[1];
  const jwtToken = localStorage.getItem('jwtToken') || '';

  const settingsUpdate = async () => {
    const { federatedSearchable, followAutoAccept, searchableLocally, searchableToRobots, nsfwBlur } = getSettingsElements();
    if (federatedSearchable && followAutoAccept && searchableLocally && searchableToRobots && nsfwBlur ) {

      const res = await axios.patch(`/api/members/update/${memberName}/preferences`, settings, {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwtToken}`,
          'X-CSRF-Token': csrfToken || ''
        }
      });  
      if (res.status !== 200) {
        errorMessages.push(`Error updating settings: ${res.data.message} (${res.status})`);
        errorMessages = [...errorMessages]
      }
      settingsSaved = true;
      
      dispatch('privacySettingsUpdated', {
        newSettings: settings 
      });
    } 
  };
</script>
<form id="privacy-settings" on:submit={settingsUpdate}>
<h3 class="settings-section-descriptor">{$_('interactions')}</h3>
  <label class="settings-label" for="auto-accept-follow">{$_('auto_accept_follow')}
</label>
<input type="checkbox" id="auto-accept-follow" bind:value={settings.auto_accept_follow} />
<div class="settings-text-input">
  <label for="muted-instances">{$_('muted')} {$_('instances')}</label>
  <label for="confirm-muting-instance">{$_('require_instance_mute_confirmation')}</label>
  <input type="checkbox" bind:value={confirmMutingInstance}/>
  {#if confirmMutingInstance}
  <Tags
    bind:tags={settings.muted_instances}
    onlyUnique={true}
    autoComplete={knownInstances}
    onlyAutoComplete={true}
    onTagAdded={confirm('Really mute this instance?')}
    />
    {:else}
      <!-- TODO: cosider using element.setAttribute -->
  <Tags
    bind:tags={settings.muted_instances}
    onlyUnique={true}
    autoComplete={knownInstances}
    onlyAutoComplete={true}
    />
    {/if}
   <label for="autohide-words">
    {$_('message_autohide_words')}
   </label> 
    <Tags
      bind:tags={settings.message_autohide_words}
      onlyUnique={true}
      />
</div>

<hr />

<h3 class="settings-section-descriptor">{$_('who_can_search_my_profile')}</h3>

<label class="settings-label" for="federated searchable">
  {$_('known_network')}
</label>
<input type="checkbox" id="federated-searchable" bind:value={settings.searchable_to_federated}/>

<label class="settings-label" for="locally-searchable">
  {$_('local_accounts')}
  </label> 
<input type="checkbox" id="locally-searchable" bind:value={settings.locally_searchable}/>

<label class="settings-label" for="robots-searchable">
  {$_('searchable_to_robots')}
</label>
<input type="checkbox" id="robots-searchable" bind:value={settings.robots_searchable} />

<hr />
</form>