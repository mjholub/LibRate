<script defer lang="ts">
  import { _ } from 'svelte-i18n';
  import { fade } from 'svelte/transition'; 
  import { circOut } from 'svelte/easing';

  let themeSaved = false;
	let themes: string[] = ['default', 'light', 'nord', 'solarized', 'gruvbox', 'dracula'];
  let theme = localStorage.getItem('theme') || 'default';

  function setTheme() {
    themeSaved = true;
    localStorage.setItem('theme', theme);
 }
 function unsetTheme() {
    if (theme !== 'default') {
      localStorage.removeItem('theme');
    }
 }

</script>
{#if theme === 'light'}
<style>
  :root {
    --main-bg-color: #faeffe !important;
    --body-bgcolor: #faeffe !important;
    --text-color: #111 !important;
    --member-card-background-color: ghostwhite !important;			
  }
</style>
{/if}


<label for="theme-selector">
  {$_('theme')}:
</label>
<select id="theme-selector" bind:value={theme}>
  {#each themes as theme}
    <option value={theme}>{theme}</option>
  {/each}
</select>
{#if themeSaved}
<p transition:fade={{ delay: 500, duration: 500, easing: circOut }} class="confirmation-message">
{$_('settings')} {$_('saved')}
</p>
{/if}
<div class="save-reset-buttons">
  <button type="submit" class="save-button" on:click={setTheme}>
  {$_('save')}
</button>
<button type="submit" class="clear-button" on:click={unsetTheme}>
{$_('reset')}
</button>
</div>


<style>
  #theme-selector {
 margin-bottom: 1em;	
}
  
button[type="submit"] {
  border-radius: var(--button-minor-border-radius);
}
</style>